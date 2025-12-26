package routes

import (
	"MuXi/Library/controller"
	"MuXi/Library/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", controller.Register)
			auth.POST("/login", controller.Login)
		}

		authGroup := api.Group("/")
		authGroup.Use(middleware.AuthRequired())
		{
			authGroup.POST("/logout", controller.Logout)
			borrows := authGroup.Group("/borrows")
			{
				borrows.POST("", controller.BorrowBook)
				borrows.POST("/return", controller.ReturnBook)
			}

			authGroup.GET("/books", controller.GetBooks)

			adminGroup := authGroup.Group("/")
			adminGroup.Use(middleware.AdminRequired())
			{
				adminGroup.POST("/books", controller.CreateBook)
				adminGroup.PUT("/books/:book_id", controller.UpdateBook)
				adminGroup.DELETE("/books/:book_id", controller.DeletedBook)
			}
		}
	}
}
