package domain

import "time"

type User struct {
	ID           uint
	Username     string
	DisplayName  string
	Email        string
	PasswordHash string
	Avatar       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
