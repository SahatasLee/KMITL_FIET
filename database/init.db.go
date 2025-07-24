package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/microsoft/go-mssqldb"
)

func DatabaseInit() *sqlx.DB {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	server := os.Getenv("DB_SERVER")
	port := os.Getenv("DB_PORT")
	database := os.Getenv("DB_DATABASE")

	if user == "" || password == "" || server == "" || port == "" || database == "" {
		log.Fatal("Database environment variables are not fully set")
	}

	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s&encrypt=disable",
		user, password, server, port, database)

	db, err := sqlx.Connect("sqlserver", dsn)
	for i := 0; i < 10; i++ {
		if err == nil {
			break
		}
		log.Printf("Retries conneting database... (%d/10)", i+1)
		time.Sleep(10 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	log.Println("Connected to SQL Server")
	return db
}
