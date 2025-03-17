package router

import (
	"fiet/controller"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func SetUserRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	ctls := controller.DBController{Database: db}
	// router.POST("/register", ctls.Register)
	// router.POST("/login", ctls.Login)
	router.GET("/users/:id", ctls.GetUserByID)
	router.GET("/users", ctls.GetUsers)
	router.PATCH("/users", ctls.UpdateUser)
	router.DELETE("/users/:id", ctls.DeleteUserByID)
}
