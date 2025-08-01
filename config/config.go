package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadConfig() {
	// Load environment variables from .env file
	err := godotenv.Load("dev.env")
	if err != nil {
		log.Println("Warning loading .env file")
	}
	fmt.Println("JWT secret", os.Getenv("JWT_SECRET"))
}
