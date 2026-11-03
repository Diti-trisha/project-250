package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var DB *pgxpool.Pool

func ConnectDB() error {
	// load .env file
	_ = godotenv.Load()

	// load db url from os environment variables
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return fmt.Errorf("DB_URL environment variable is not set.")
	}

	// connect to the database
	var err error
	DB, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %v", err)
	}

	// verify connection
	if err := DB.Ping(context.Background()); err != nil {
		return fmt.Errorf("Database ping failed: %v", err)
	}

	fmt.Println("Successfully connected to the database!")
	return nil
}

// execute schema.sql
func RunMigration(filePath string) error {
	sqlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Failed to read schema file: %w", err)
	}

	_, err = DB.Exec(context.Background(), string(sqlBytes))
	if err != nil {
		return fmt.Errorf("Failed to execute schema migration: %w", err)
	}

	fmt.Println("Database migration executed successfully!")
	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
