package router

import (
	"fiet/controller"
	"fiet/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func SetUserRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	ctls := controller.DBController{Database: db}

	// Public routes
	router.POST("/register", ctls.CreateUser)
	router.POST("/login", ctls.Login)

	// Protected routes with middleware
	protected := router.Group("/")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		protected.GET("/users", ctls.GetUsers)
		protected.GET("/user", ctls.GetUserByID)
		protected.PATCH("/user", ctls.UpdateUser)
		protected.DELETE("/user", ctls.DeleteUserByID)
	}
}
