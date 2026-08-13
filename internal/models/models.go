package models

import "time"

type User struct {
	ID             int64
	Username       string
	PasswordHash   string
	FailedAttempts int
	LockedUntil    *time.Time
	LastLoginAt    *time.Time
	CreatedAt      time.Time
}

type Session struct {
	ID        int64
	UserID    int64
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type TOTPSecret struct {
	UserID    int64
	Secret    string
	Enabled   bool
	CreatedAt time.Time
}
