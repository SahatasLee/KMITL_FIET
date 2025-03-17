package db

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
)

func DatabaseInit() *sqlx.DB {
	// database connection
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	server := os.Getenv("DB_SERVER")
	port := os.Getenv("DB_PORT")
	database := os.Getenv("DB_DATABASE")

	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s&encrypt=disable",
		user, password, server, port, database)

	db, err := sqlx.Connect("sqlserver", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	return db
}
