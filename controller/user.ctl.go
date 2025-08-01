package controller

import (
	"database/sql"
	"fiet/auth"
	"fiet/model"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// @Summary      Create User
// @Description  Create a new user
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        credentials  body     model.Credential  true  "User Credentials"
// @Success      201  {string}  "User created successfully"
// @Failure      400  {string}  model.ErrorResponse
// @Failure      409  {string}  "User already exists"
// @Failure      500  {string}  model.ErrorResponse
// @Router       /register [post]
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email, and password are required"})
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
		INSERT INTO users (uuid, email, password_hash)
		OUTPUT INSERTED.id
		VALUES (:uuid, :email, :password_hash)
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

// Login user
// @Summary      Login User
// @Description  Authenticate user and return JWT token
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        credentials  body     model.Credential  true  "User Credentials"
// @Success	  	 200  {object}	model.TokenResponse "Successful login"
// @Failure      400  {string}  "Invalid input"
// @Failure      401  {string}  "Invalid email or password"
// @Failure      500  {string}  "Internal server error"
// @Router       /login [post]
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
	query := "SELECT uuid, password_hash FROM users WHERE email = :email"
	stmt, err := db.Database.PrepareNamed(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query preparation failed"})
		return
	}
	defer stmt.Close()

	if err := stmt.Get(&user, map[string]interface{}{"email": req.Email}); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password."})
		c.Error(err)
		return
	}

	// Compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		c.Error(err)
		return
	}

	// uuid := FixUUIDFromSQLServer(user.UUID)
	uuid := user.UUID

	fmt.Printf("Extracted UUID from JWT: %v (%T)\n", uuid, uuid)
	// Success response (excluding password)
	token, err := auth.GenerateToken(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}
	fmt.Println("Generated JWT token:", token)
	c.SetCookie(
		"token",     // name
		token,       // value
		3600,        // maxAge in seconds (e.g., 1 hour)
		"/",         // path
		"localhost", // domain — use frontend domain
		false,       // secure (true = HTTPS only)
		false,       // httpOnly (JS can't access it)
	)
	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
	// c.JSON(http.StatusOK, gin.H{"token": token})
}

// Get all users
// TODO: Implement pagination and filtering
// @Summary      Get Users
// @Description  Retrieve all users
// @Tags         user
// @Accept       json
// @Produce      json
// @Success      200  {array}   model.PublicUser
// @Failure      500  {string}  "Internal server error"
// @Router       /users [get]
// @Security 	 BearerAuth
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
// @Summary      Get User by UUID
// @Description  Retrieve user details by UUID from JWT
// @Tags         user
// @Accept       json
// @Produce      json
// @Success      200  {object}  model.PublicUser
// @Failure      401  {string}  "Unauthorized"
// @Failure      404  {string}  "User not found"
// @Failure      500  {string}  "Internal server error"
// @Router       /user [get]
// @Security 	 BearerAuth
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

// Update user by UUID from JWT
// @Summary      Update User
// @Description  Update user details by UUID from JWT
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        user  body     model.PublicUser true  "User details to update"
// @Success      200  {string}  "User updated successfully"
// @Failure      400  {string}  "Invalid request data"
// @Failure      401  {string}  "Unauthorized"
// @Failure      404  {string}  "User not found"
// @Failure      500  {string}  "Failed to update user"
// @Router       /user [patch]
// @Security 	 BearerAuth
func (db *DBController) UpdateUser(c *gin.Context) {
	// Extract user UUID from JWT
	userUUIDVal, exists := c.Get("user_uuid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User UUID not found"})
		return
	}
	userUUID := userUUIDVal.(string)

	// Parse incoming JSON into a map
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Whitelist allowed fields
	allowedFields := map[string]bool{
		"name":  true,
		"age":   true,
		"email": true,
		// Add more if needed
	}

	// Build SQL SET clause
	setClauses := []string{}
	params := map[string]interface{}{
		"uuid": userUUID,
	}

	for key, value := range req {
		if allowedFields[key] {
			setClauses = append(setClauses, fmt.Sprintf("%s = :%s", key, key))
			params[key] = value
		}
	}

	if len(setClauses) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	// Always update updated_at
	setClauses = append(setClauses, "updated_at = SYSDATETIME()")

	query := fmt.Sprintf("UPDATE users SET %s WHERE uuid = :uuid", strings.Join(setClauses, ", "))
	result, err := db.Database.NamedExec(query, params)
	if err != nil {
		log.Println("Update error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

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

// Change user password
func (db *DBController) ChangePassword(c *gin.Context) {
	// Change user password
	type ChangePasswordRequest struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	// Get user UUID from JWT
	userUUIDVal, exists := c.Get("user_uuid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userUUID := userUUIDVal.(string)

	// Parse JSON input
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Fetch existing password hash
	var storedHash string
	err := db.Database.Get(&storedHash, "SELECT password_hash FROM users WHERE uuid = ?", userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	// Compare current password with stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect current password"})
		return
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	// Update password in DB
	_, err = db.Database.Exec("UPDATE users SET password_hash = ?, updated_at = SYSDATETIME() WHERE uuid = ?", newHash, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
