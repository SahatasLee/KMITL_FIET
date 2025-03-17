package controller

import (
	"fiet/model"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

// Create a new user
func (db *DBController) CreateUser(c *gin.Context) {
	var user model.User

	// Bind JSON request to struct
	if err := c.Bind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "INSERT INTO users (name, age) OUTPUT INSERTED.id VALUES (:name, :age)"

	// Prepare query
	stmt, err := db.Database.PrepareNamed(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare query"})
		return
	}
	defer stmt.Close()

	// Execute query and get the inserted ID
	var id int64
	err = stmt.Get(&id, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert user"})
		fmt.Println("Insert Error:", err)
		return
	}

	// Send success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"id":      id,
	})
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
