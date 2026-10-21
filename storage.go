package main

import (
	"sync"
)

type User struct {
	RegNo    int    `json:"reg_no"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Dept     string `json:"dept"`
	PhoneNo  string `json:"phone_no"`
}

type Storage struct {
	mu    sync.RWMutex
	users map[int]User
}

func NewStorage() *Storage {
	return &Storage{
		users: make(map[int]User),
	}
}

func (s *Storage) Register(u User) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[u.RegNo]; exists {
		return false
	}

	s.users[u.RegNo] = u
	return true
}

func (s *Storage) Validate(regNo int, password string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[regNo]
	return exists && user.Password == password
}

func (s *Storage) Get(regNo int) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[regNo]
	return user, exists
}
