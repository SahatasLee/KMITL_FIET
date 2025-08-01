package auth

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var jwtSecret []byte // 🔐 Store this securely (env or config)

func init() {
	fmt.Println("JWT Secret:", jwtSecret)
	err := godotenv.Load("dev.env")
	if err != nil {
		log.Println("Warning loading .env file")
	}
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	fmt.Println("JWT Secret:", jwtSecret)
}

func GenerateToken(userUUID string) (string, error) {
	claims := jwt.MapClaims{
		"user_uuid": userUUID,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
		"iat":       time.Now().Unix(), // issued at
		"nbf":       time.Now().Unix(), // not before
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
}
