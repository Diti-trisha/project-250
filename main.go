package main

import (
	"context"
	"log"
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
	ctx := context.Background()

	// test: create user
	newUser := &User{
		RegNo:    2023331027,
		FullName: "DITI BHATTACHARJEE",
		Email:    "2023331027@student.sust.edu",
		Password: "abcde",
		Dept:     "CSE",
		PhoneNo:  "01716842832",
	}

	_ = userRepo.CreateUser(ctx, newUser)

	// test: search user
	fetchedUser, err := userRepo.GetUserByRegNo(ctx, 2023331027)
	if err != nil {
		log.Fatalf("Error fetching user: %v", err)
	}

	log.Printf("Found User: %+v\n", fetchedUser)
}
