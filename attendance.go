package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StudentRow struct {
	RegNo    int    `json:"reg_no"`
	FullName string `json:"full_name"`
}

type AttendanceRecord struct {
	RegNo   int    `json:"reg_no"`
	Date    string `json:"date"` // dd-mm-yyyy
	Present bool   `json:"present"`
}

type AttendanceSheetResponse struct {
	Dates    []string           `json:"dates"`
	Students []StudentRow       `json:"students"`
	Records  []AttendanceRecord `json:"records"`
}

type ToggleAttendanceRequest struct {
	RegNo   int    `json:"reg_no"`
	Date    string `json:"date"` // dd-mm-yyyy
	Present bool   `json:"present"`
}

type AttendanceHandler struct {
	db *pgxpool.Pool
}

func NewAttendanceHandler(db *pgxpool.Pool) *AttendanceHandler {
	return &AttendanceHandler{db: db}
}

// get /api/attendance
func (h *AttendanceHandler) GetSheet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role, _ := r.Context().Value("role").(string)
	if strings.ToLower(role) != "teacher" {
		http.Error(w, "Forbidden: teacher access required", http.StatusForbidden)
		return
	}

	const numDays = 14
	now := time.Now().In(utcPlus6)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, utcPlus6)
	startDate := today.AddDate(0, 0, -(numDays - 1))
	endExclusive := today.AddDate(0, 0, 1)

	dates := make([]string, 0, numDays)
	for d := startDate; d.Before(endExclusive); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format(dateLayout))
	}

	studentQuery := `
		SELECT reg_no, full_name
		FROM users
		WHERE LOWER(role) = 'student'
		ORDER BY reg_no ASC
	`
	studentRows, err := h.db.Query(r.Context(), studentQuery)
	if err != nil {
		http.Error(w, "Failed to query students", http.StatusInternalServerError)
		return
	}
	defer studentRows.Close()

	students := make([]StudentRow, 0)
	for studentRows.Next() {
		var s StudentRow
		if err := studentRows.Scan(&s.RegNo, &s.FullName); err != nil {
			http.Error(w, "Failed to read student row", http.StatusInternalServerError)
			return
		}
		students = append(students, s)
	}
	if err := studentRows.Err(); err != nil {
		http.Error(w, "Student rows error", http.StatusInternalServerError)
		return
	}

	attQuery := `
		SELECT reg_no, attendance_date, present
		FROM attendance
		WHERE attendance_date >= $1 AND attendance_date < $2
		ORDER BY attendance_date ASC, reg_no ASC
	`
	attRows, err := h.db.Query(r.Context(), attQuery, startDate, endExclusive)
	if err != nil {
		http.Error(w, "Failed to query attendance", http.StatusInternalServerError)
		return
	}
	defer attRows.Close()

	records := make([]AttendanceRecord, 0)
	for attRows.Next() {
		var regNo int
		var attDate time.Time
		var present bool
		if err := attRows.Scan(&regNo, &attDate, &present); err != nil {
			http.Error(w, "Failed to read attendance row", http.StatusInternalServerError)
			return
		}
		records = append(records, AttendanceRecord{
			RegNo:   regNo,
			Date:    attDate.Format(dateLayout),
			Present: present,
		})
	}
	if err := attRows.Err(); err != nil {
		http.Error(w, "Attendance rows error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(AttendanceSheetResponse{
		Dates:    dates,
		Students: students,
		Records:  records,
	})
}

// put /api/attendance
func (h *AttendanceHandler) SetTick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role, _ := r.Context().Value("role").(string)
	if strings.ToLower(role) != "teacher" {
		http.Error(w, "Forbidden: teacher access required", http.StatusForbidden)
		return
	}

	var req ToggleAttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.RegNo <= 0 {
		http.Error(w, "Invalid reg_no", http.StatusBadRequest)
		return
	}

	parsedDate, err := time.Parse(dateLayout, strings.TrimSpace(req.Date))
	if err != nil {
		http.Error(w, "Date must be DD-MM-YYYY", http.StatusBadRequest)
		return
	}
	targetDate := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, utcPlus6)

	now := time.Now().In(utcPlus6)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, utcPlus6)
	if targetDate.After(today) {
		http.Error(w, "Cannot set attendance for future dates", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO attendance (reg_no, attendance_date, present)
		VALUES ($1, $2, $3)
		ON CONFLICT (reg_no, attendance_date) DO UPDATE
		SET present = EXCLUDED.present
	`
	if _, err := h.db.Exec(r.Context(), query, req.RegNo, targetDate, req.Present); err != nil {
		http.Error(w, "Failed to save attendance", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(AttendanceRecord{
		RegNo:   req.RegNo,
		Date:    targetDate.Format(dateLayout),
		Present: req.Present,
	})
}