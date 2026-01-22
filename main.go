//	@title			MuXiShooter
//	@version		1.0
//	@description	MuXiShooter
//	@termsOfService	http://swagger.io/terms
//	@contact.name	FireChickenMP4
//	@contact.email	13930176445@163.com
//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT
//	@host			localhost:8080
//	@BasePath		/api

package main

import (
	config "MuXi/2026-MuxiShooter-Backend/config"
	_ "MuXi/2026-MuxiShooter-Backend/docs"
	"MuXi/2026-MuxiShooter-Backend/dto"
	"MuXi/2026-MuxiShooter-Backend/middleware"
	routes "MuXi/2026-MuxiShooter-Backend/routes"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	config.ConnectDB()
	config.InitAdmin(config.DB)
	config.InitJWTSecret()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{ // 允许的请求源
			"http://localhost:5173", // 前端vite的默认启动地址
			"http://localhost:3000", // 前端自己定义的启动地址
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},                   // 允许的请求方法
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"}, // 允许的请求头
		AllowCredentials: true,
		MaxAge:           1 * time.Hour,
	}))
	r.Use(middleware.Limiter())
	r.Use(func(c *gin.Context) {
		c.Next()
		if c.Writer.Status() == http.StatusTooManyRequests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.Response{
				Code:    http.StatusTooManyRequests,
				Message: "请求过于频繁，请稍后重试(1s)",
			})
		}
	})
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	//使用gzip传输

	r.Static("/uploads", "./uploads")
	//gin的Static是Gin框架中用来提供静态文件服务的功能，就像在餐厅里设置一个自助区
	//让顾客可以自己取用饮料和小食，而不需要每次都找服务员点单。

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	routes.RegisterRoutes(r)

	// test.TestReferenceTableWithDB(config.DB)
	// test.CleanTestData(config.DB)

	log.Println("服务器启动在 http://localhost:8080")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器启动失败: %v\n", err)
	}
}
