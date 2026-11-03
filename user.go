package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	RegNo    int    `json:"reg_no"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Dept     string `json:"dept"`
	PhoneNo  string `json:"phone_no"`
}

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// insert a new user into the database
func (r *UserRepository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (reg_no, full_name, email, password, dept, phone_no)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		user.RegNo,
		user.FullName,
		user.Email,
		user.Password,
		user.Dept,
		user.PhoneNo,
	)
	if err != nil {
		return fmt.Errorf("Failed to insert user: %v", err)
	}
	return nil
}

// fetch a user by their registration number
func (r *UserRepository) GetUserByRegNo(ctx context.Context, regNo int) (*User, error) {
	query := `
		SELECT reg_no, full_name, email, password, dept, phone_no
		FROM users WHERE reg_no = $1
	`
	var user User
	err := r.db.QueryRow(ctx, query, regNo).Scan(
		&user.RegNo,
		&user.FullName,
		&user.Email,
		&user.Password,
		&user.Dept,
		&user.PhoneNo,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("User with reg no %d not found", regNo)
		}
		return nil, fmt.Errorf("Failed to query user: %w", err)
	}

	return &user, nil
}
