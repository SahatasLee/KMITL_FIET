package model

import (
	"database/sql"
	"time"
)

type User struct {
	ID        int            `db:"id" json:"-"`      // Internal ID (never exposed)
	UUID      string         `db:"uuid" json:"uuid"` // Public-safe ID
	Name      sql.NullString `db:"name" json:"name"`
	Email     string         `db:"email" json:"email"`
	Age       sql.NullInt64  `db:"age" json:"age"`
	Password  string         `db:"password_hash" json:"password"` // Hashed password
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
}

type PublicUser struct {
	UUID      string         `db:"uuid" json:"uuid"`
	Name      sql.NullString `db:"name" json:"name"`
	Email     string         `db:"email" json:"email"`
	Age       sql.NullInt64  `db:"age" json:"age"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
}
