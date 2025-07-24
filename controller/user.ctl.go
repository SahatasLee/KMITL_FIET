package controller

import (
	"database/sql"
	"fiet/auth"
	"fiet/model"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (db *DBController) CreateUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	// Bind JSON with validation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Validate required fields manually if needed
	if req.Email == "" || req.Password == "" {
		fmt.Println(req.Email, req.Password)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name, email, and password are required"})
		return
	}

	// Check if user already exists
	query := "SELECT id FROM users WHERE email = :email"

	stmt, err := db.Database.PrepareNamed(query)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query preparation failed"})
		return
	}
	defer stmt.Close()

	var existingID int
	err = stmt.Get(&existingID, gin.H{"email": req.Email})

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	} else if err != sql.ErrNoRows {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Generate UUID
	newUUID := uuid.New().String()

	// Hash password securely (bcrypt)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hashing failed"})
		return
	}

	// Insert query with named params
	query = `
		INSERT INTO users (uuid, name, email, age, password_hash)
		OUTPUT INSERTED.id
		VALUES (:uuid, :name, :email, :age, :password_hash)
	`

	// Prepare statement
	stmt, err = db.Database.PrepareNamed(query)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query preparation failed"})
		return
	}
	defer stmt.Close()

	// Params to bind
	params := map[string]interface{}{
		"uuid":          newUUID,
		"email":         req.Email,
		"password_hash": string(hashedPassword),
	}

	// Execute and fetch new ID
	var id int
	if err := stmt.Get(&id, params); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User creation failed"})
		c.Error(err)
		// log.Println("DB Error:", err)
		return
	}

	// Success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created",
	})
}

func (db *DBController) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	// Bind JSON with validation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Fetch user by email
	var user model.User
	query := "SELECT id, uuid, name, email, password_hash FROM users WHERE email = :email"
	stmt, err := db.Database.PrepareNamed(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query preparation failed"})
		return
	}
	defer stmt.Close()

	if err := stmt.Get(&user, map[string]interface{}{"email": req.Email}); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Success response (excluding password)
	token, err := auth.GenerateToken(user.UUID) // assume user ID is 1
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Get all users
func (db *DBController) GetUsers(c *gin.Context) {
	var users []model.User
	query := "SELECT id, name, age FROM users"

	err := db.Database.SelectContext(c.Request.Context(), &users, query)
	if err != nil {
		log.Println("Error fetching users:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Send the user list as a response
	c.JSON(http.StatusOK, users)
}

// Get user by ID
func (db *DBController) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	query := "SELECT id, name, age FROM users WHERE id=:id"

	var user model.User
	stmt, err := db.Database.PrepareNamed(query)
	if err != nil {
		log.Println("Error preparing query:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare query"})
		return
	}
	defer stmt.Close()

	// Execute the query
	err = stmt.Get(&user, map[string]interface{}{"id": id})
	if err != nil {
		log.Println("Error fetching user:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Send the user details as a response
	c.JSON(http.StatusOK, user)
}

// Update user by ID
func (db *DBController) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	// Bind JSON body to struct
	var user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	query := "UPDATE users SET name=:name, age=:age WHERE id=:id"
	params := map[string]interface{}{
		"id":   id,
		"name": user.Name,
		"age":  user.Age,
	}

	// Execute update query
	result, err := db.Database.NamedExec(query, params)
	if err != nil {
		log.Println("Update error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Check if any rows were affected
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Send success response
	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// Delete user by ID
func (db *DBController) DeleteUserByID(c *gin.Context) {
	// Get user ID from URL parameter
	id := c.Param("id")

	query := "DELETE FROM dbo.users WHERE id = :id"

	// Execute delete query
	result, err := db.Database.NamedExec(query, map[string]interface{}{"id": id})
	if err != nil {
		log.Println("Delete error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	// Check if any rows were actually deleted
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Send success response
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
