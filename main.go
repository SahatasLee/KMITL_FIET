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
	// Trust a specific proxy (e.g., NGINX running on 10.0.0.1)
	r.SetTrustedProxies([]string{"10.0.0.1", "192.168.1.0/24", "localhost"})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "hello world",
		})
	})
	api := r.Group("/api/v1")
	router.SetUserRoutes(api, db)
	r.Run(":80") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
