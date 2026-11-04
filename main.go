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

	// initialize user repo
	userRepo := NewUserRepository(DB)
	authHandler := NewAuthHandler(userRepo)

	// endpoints
	http.HandleFunc("/api/register", authHandler.Register)
	http.HandleFunc("/api/login", authHandler.Login)
	http.HandleFunc("/api/profile", JWTMiddleware(authHandler.ProtectedProfile))

	log.Println("Server running on http://localhost:8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
