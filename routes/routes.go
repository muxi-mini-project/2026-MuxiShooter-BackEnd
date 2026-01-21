package routes

import (
	"MuXi/2026-MuxiShooter-Backend/controller"
	"MuXi/2026-MuxiShooter-Backend/middleware"

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
		authGroup.Use(middleware.JWTAuth())
		{
			profile := authGroup.Group("/profile")
			{
				update := profile.Group("/update")
				{
					update.PUT("/password", controller.UpdatePassword)
					update.PUT("/username", controller.UpdateUsername)
					update.PUT("/headimage", controller.UpdateHeadImage)
				}
			}

			adminGroup := authGroup.Group("/")
			adminGroup.Use(middleware.AdminRequired())
			{
			}
		}
	}
}
