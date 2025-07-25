package model

import (
	"time"
)

type User struct {
	ID        int       `db:"id" json:"-"`      // Internal ID (never exposed)
	UUID      string    `db:"uuid" json:"uuid"` // Public-safe ID
	Name      *string   `db:"name" json:"name,omitempty"`
	Email     string    `db:"email" json:"email"`
	Age       *int64    `db:"age" json:"age,omitempty"`
	Password  string    `db:"password_hash" json:"password"` // Hashed password
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type PublicUser struct {
	UUID      string    `db:"uuid" json:"uuid"`
	Name      *string   `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Age       *int64    `db:"age" json:"age"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Credential struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8,max=64" example:"supersecure123"`
}

type TokenResponse struct {
	Token string `json:"token"`
}
