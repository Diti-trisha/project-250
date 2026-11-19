package main

import (
	"log"
	"net/http"
)

func main() {
	// connect to database
	if err := ConnectDB(); err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer CloseDB()

	// run initial schema
	if err := RunMigration("schema.sql"); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	// initialize repos and handlers
	userRepo := NewUserRepository(DB)
	authHandler := NewAuthHandler(userRepo)
	calendarHandler := NewCalendarHandler(DB)

	// frontend pages
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "login.html")
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "register.html")
	})
	http.HandleFunc("/calendar", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "calendar.html")
	})

	// api endpoints
	http.HandleFunc("/api/register", authHandler.Register)
	http.HandleFunc("/api/login", authHandler.Login)
	http.HandleFunc("/api/calendar", JWTMiddleware(calendarHandler.GetMonth))
	http.HandleFunc("/api/calendar/day", JWTMiddleware(calendarHandler.SetDay))

	log.Println("Server running on http://localhost:8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}