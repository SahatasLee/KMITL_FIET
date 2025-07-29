package main

import (
	"fiet/config"
	db "fiet/database"
	docs "fiet/docs"
	"fiet/router"
	"net/http"

	"github.com/gin-contrib/cors"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

// @title           Fiet API
// @version         1.0
// @description     This is Fiet.

// @host      localhost:8080
// @BasePath  /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Authorization: Bearer {token}"
func main() {
	// Load environment variables from .env file
	config.LoadConfig()

	// Initialize the database connection
	db := db.DatabaseInit()

	r := gin.Default()
	// Trust a specific proxy (e.g., NGINX running on 10.0.0.1)
	r.SetTrustedProxies([]string{"10.0.0.1", "192.168.1.0/24", "localhost"})
	r.Use(cors.Default())
	docs.SwaggerInfo.BasePath = "/api/v1"

	api := r.Group("/api/v1")

	api.GET("/ping", PingHandler)

	router.SetUserRoutes(api, db)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	r.Run(":8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

// Ping godoc
// @Summary      ping
// @Description  do ping
// @Tags         ping
// @Accept       json
// @Produce      json
// @Success      200  {string}  "hello world"
// @Router       /ping [get]
func PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "hello world",
	})
}
