package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const dateLayout = "02-01-2006"

type DayType string

const (
	DayRegular DayType = "regular"
	DayHoliday DayType = "holiday"
	DayEvent   DayType = "event"
)

type CalendarDay struct {
	Date           string  `json:"date"`            // dd-mm-yyyy
	Type           DayType `json:"type"`            // regular | holiday | event
	DayDescription string  `json:"day_description"` // holiday or event description
}

type SetDayRequest struct {
	Date           string  `json:"date"`            // dd-mm-yyyy
	Type           DayType `json:"type"`            // regular | holiday | event
	DayDescription string  `json:"day_description"` // holiday or event description
}

type CalendarHandler struct {
	db *pgxpool.Pool
}

func NewCalendarHandler(db *pgxpool.Pool) *CalendarHandler {
	return &CalendarHandler{db: db}
}

// get /api/calendar?month=01&year=2026
func (h *CalendarHandler) GetMonth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	now := time.Now()
	utcPlus6 := time.FixedZone("UTC+6", 6*60*60)
	year, err := parseIntQuery(r.URL.Query().Get("year"), now.In(utcPlus6).Year())
	if err != nil || year < 1970 || year > 2100 {
		http.Error(w, "Invalid year", http.StatusBadRequest)
		return
	}

	month, err := parseIntQuery(r.URL.Query().Get("month"), int(now.In(utcPlus6).Month()))
	if err != nil || month < 1 || month > 12 {
		http.Error(w, "Invalid month (1-12)", http.StatusBadRequest)
		return
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, utcPlus6)
	endDate := startDate.AddDate(0, 1, 0)

	query := `
		SELECT calendar_date, day_type, day_description
		FROM calendar_days
		WHERE calendar_date >= $1 AND calendar_date < $2
		ORDER BY calendar_date ASC
	`
	rows, err := h.db.Query(r.Context(), query, startDate, endDate)
	if err != nil {
		http.Error(w, "Failed to query calendar", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	days := make([]CalendarDay, 0)
	for rows.Next() {
		var d time.Time
		var dayType string
		var desc string
		if err := rows.Scan(&d, &dayType, &desc); err != nil {
			http.Error(w, "Failed to read calendar row", http.StatusInternalServerError)
			return
		}
		days = append(days, CalendarDay{
			Date:           d.Format(dateLayout),
			Type:           DayType(dayType),
			DayDescription: desc,
		})
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Calendar rows error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"year":  year,
		"month": month,
		"days":  days,
	})
}

// put /api/calendar/day
func (h *CalendarHandler) SetDay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role, _ := r.Context().Value("role").(string)
	if strings.ToLower(role) != "admin" {
		http.Error(w, "Forbidden: admin access required", http.StatusForbidden)
		return
	}

	var req SetDayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	parsedDate, err := time.Parse(dateLayout, strings.TrimSpace(req.Date))
	if err != nil {
		http.Error(w, "Date must be DD-MM-YYYY", http.StatusBadRequest)
		return
	}

	req.Type = DayType(strings.ToLower(strings.TrimSpace(string(req.Type))))
	req.DayDescription = strings.TrimSpace(req.DayDescription)

	if req.Type != DayRegular && req.Type != DayHoliday && req.Type != DayEvent {
		http.Error(w, "Type must be regular, holiday, or event", http.StatusBadRequest)
		return
	}

	if req.Type == DayRegular {
		req.DayDescription = ""
	} else if req.DayDescription == "" {
		http.Error(w, "Description is required for holidays and events", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO calendar_days (calendar_date, day_type, day_description)
		VALUES ($1, $2, $3)
		ON CONFLICT (calendar_date) DO UPDATE
		SET day_type = EXCLUDED.day_type,
		    day_description = EXCLUDED.day_description
	`
	if _, err := h.db.Exec(r.Context(), query, parsedDate, string(req.Type), req.DayDescription); err != nil {
		http.Error(w, "Failed to save calendar day", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(CalendarDay{
		Date:           parsedDate.Format(dateLayout),
		Type:           req.Type,
		DayDescription: req.DayDescription,
	})
}

func parseIntQuery(raw string, fallback int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return fallback, nil
	}
	return strconv.Atoi(raw)
}