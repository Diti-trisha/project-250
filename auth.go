package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repo *UserRepository
}

func NewAuthHandler(repo *UserRepository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

type Claims struct {
	RegNo int `json:"reg_no"`
	jwt.RegisteredClaims
}

// handle user signup
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload.", http.StatusBadRequest)
		return
	}

	// hash password securely
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	// save user to database
	ctx := context.Background()
	if err := h.repo.CreateUser(ctx, &user); err != nil {
		http.Error(w, "Failed to create user.", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

// handle user authentication and return jwt token
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		RegNo    int    `json:"reg_no"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	user, err := h.repo.GetUserByRegNo(ctx, creds.RegNo)
	if err != nil {
		http.Error(w, "Invalid registration number or password.", http.StatusUnauthorized)
		return
	}

	// verify hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		http.Error(w, "Invalid registration number or password.", http.StatusUnauthorized)
		return
	}

	// generate jwt token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		RegNo: user.RegNo,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, "Could not generate token.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// handle an authenticated route request
func (h *AuthHandler) ProtectedProfile(w http.ResponseWriter, r *http.Request) {
	regNo := r.Context().Value("reg_no").(int)

	ctx := context.Background()
	user, err := h.repo.GetUserByRegNo(ctx, regNo)
	if err != nil {
		http.Error(w, "User not found.", http.StatusNotFound)
		return
	}

	// clear password before sending data back
	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
