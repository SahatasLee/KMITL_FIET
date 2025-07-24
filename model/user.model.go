package model

import (
	"time"
)

type User struct {
	ID        int       `db:"id" json:"-"`      // Internal ID (never exposed)
	UUID      []byte    `db:"uuid" json:"uuid"` // Public-safe ID
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Age       int       `db:"age" json:"age"`
	Password  string    `db:"password_hash" json:"password"` // Hashed password
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type PublicUser struct {
	UUID      []byte    `db:"uuid" json:"uuid"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Age       int       `db:"age" json:"age"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
