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

func FixUUIDFromSQLServer(b []byte) uuid.UUID {
	if len(b) != 16 {
		return uuid.Nil
	}
	// Reverse byte order for first 3 fields
	copy := make([]byte, 16)
	copy[0] = b[3]
	copy[1] = b[2]
	copy[2] = b[1]
	copy[3] = b[0]

	copy[4] = b[5]
	copy[5] = b[4]

	copy[6] = b[7]
	copy[7] = b[6]

	copy[8] = b[8]
	copy[9] = b[9]
	copy[10] = b[10]
	copy[11] = b[11]
	copy[12] = b[12]
	copy[13] = b[13]
	copy[14] = b[14]
	copy[15] = b[15]

	u, _ := uuid.FromBytes(copy)
	return u
}

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

	uuid := FixUUIDFromSQLServer(user.UUID)

	fmt.Printf("Extracted UUID from JWT: %v (%T)\n", uuid, uuid)
	// Success response (excluding password)
	token, err := auth.GenerateToken(uuid.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Get all users
// TODO: Implement pagination and filtering
func (db *DBController) GetUsers(c *gin.Context) {
	var users []model.PublicUser
	query := `
	SELECT uuid, name, email, age, created_at, updated_at
	FROM users
	`

	err := db.Database.SelectContext(c.Request.Context(), &users, query)
	if err != nil {
		log.Println("Error fetching users:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Send the user list as a response
	c.JSON(http.StatusOK, users)
}

// Get user from JWT UUID
func (db *DBController) GetUserByID(c *gin.Context) {
	// Extract user UUID from JWT claims (set by middleware)
	userUUIDVal, exists := c.Get("user_uuid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User UUID not found in token"})
		return
	}

	userUUID, ok := userUUIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid UUID format"})
		return
	}

	query := `
	SELECT uuid, name, email, age, created_at, updated_at
	FROM users
	WHERE uuid = :uuid
	`

	var user model.PublicUser
	stmt, err := db.Database.PrepareNamed(query)
	if err != nil {
		log.Println("Error preparing query:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare query"})
		return
	}
	defer stmt.Close()

	// Execute query using UUID
	err = stmt.Get(&user, map[string]interface{}{"uuid": userUUID})
	if err != nil {
		log.Println("Error fetching user by UUID:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

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
