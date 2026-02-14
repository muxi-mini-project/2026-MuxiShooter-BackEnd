package routes

import (
	"MuXi/2026-MuxiShooter-Backend/controller"
	"MuXi/2026-MuxiShooter-Backend/dto"
	"MuXi/2026-MuxiShooter-Backend/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, dto.Response{
			Code:    http.StatusOK, //200
			Message: "I'm OK.",
		})
	})
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
				operation := profile.Group("/operation")
				{
					operation.GET("/logout", controller.Logout)
				}
				get := profile.Group("/get")
				{
					get.GET("/self", controller.GetSelfProfile)
				}
			}

			adminGroup := authGroup.Group("/admin")
			adminGroup.Use(middleware.AdminRequired())
			{
				getGroup := adminGroup.Group("/get")
				{
					paginatedGroup := getGroup.Group("/")
					paginatedGroup.Use(middleware.PaginationMiddleware())
					{
						paginatedGroup.GET("/getusers", controller.GetUsers)
					}
				}
			}
		}
	}
}
