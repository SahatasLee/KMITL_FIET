package main

import (
	db "fiet/database"
	"fiet/router"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	db := db.DatabaseInit()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "hello world",
		})
	})
	api := r.Group("/api/v1")
	router.SetUserRoutes(api, db)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
