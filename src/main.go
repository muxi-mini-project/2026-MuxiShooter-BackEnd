//	@title						MuXiShooter
//	@version					1.0
//	@description				MuXiShooter
//	@termsOfService				http://swagger.io/terms
//	@contact.name				FireChickenMP4
//	@contact.email				13930176445@163.com
//	@license.name				MIT
//	@license.url				https://opensource.org/licenses/MIT
//	@host						localhost:8080
//	@BasePath					/
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				输入你的Bearer Token，格式：Bearer {token}

package main

import (
	config "MuXi/2026-MuxiShooter-Backend/config"
	"MuXi/2026-MuxiShooter-Backend/controller"
	_ "MuXi/2026-MuxiShooter-Backend/docs"
	"MuXi/2026-MuxiShooter-Backend/handler"
	"MuXi/2026-MuxiShooter-Backend/infrastructure/repository"
	"MuXi/2026-MuxiShooter-Backend/infrastructure/security"
	"MuXi/2026-MuxiShooter-Backend/middleware"
	routes "MuXi/2026-MuxiShooter-Backend/routes"
	"MuXi/2026-MuxiShooter-Backend/service"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	settings := config.LoadSettings()
	appState, err := config.Bootstrap(settings)
	if err != nil {
		log.Fatalf("应用初始化失败: %v", err)
	}

	controller.SetDB(appState.DB)
	controller.SetJWTSecret(appState.JWTSecret)
	if err := controller.ValidateDependencies(); err != nil {
		log.Fatalf("controller依赖初始化失败: %v", err)
	}

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
	// r.Use(middleware.Limiter())
	// r.Use(func(c *gin.Context) {
	// 	c.Next()
	// 	if c.Writer.Status() == http.StatusTooManyRequests {
	// 		c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.Response{
	// 			Code:    http.StatusTooManyRequests,
	// 			Message: "请求过于频繁，请稍后重试(1s)",
	// 		})
	// 	}
	// })
	//这边限流器直接扔给Caddy了，go这边不需要非常精细的限流
	//也不涉及权限组的限流
	//r.Use(gzip.Gzip(gzip.DefaultCompression))
	//使用gzip传输
	//这里用Caddy的gzip压缩就ok了

	// r.Static("/uploads", "./uploads")
	// r.Static("/static", "./static")
	//gin的Static是Gin框架中用来提供静态文件服务的功能，就像在餐厅里设置一个自助区
	//让顾客可以自己取用饮料和小食，而不需要每次都找服务员点单。
	//同样,Caddy能干这活

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	userRepository := repository.NewUserRepository(appState.DB)
	relationRepository := repository.NewRelationRepository(appState.DB)
	passwordHasher := security.NewBcryptPasswordHasher()
	tokenService := security.NewJWTTokenService(appState.JWTSecret)
	authService := service.NewAuthService(userRepository, passwordHasher, tokenService, config.DefaultHeadImagePath)
	authHandler := handler.NewAuthHandler(authService)
	profileService := service.NewProfileService(userRepository, relationRepository, passwordHasher)
	profileHandler := handler.NewProfileHandler(profileService)
	jwtAuthMiddleware := middleware.JWTAuth(tokenService, userRepository)

	routes.RegisterRoutes(r, authHandler, profileHandler, jwtAuthMiddleware)

	// test.TestReferenceTableWithDB(appState.DB)
	// test.CleanTestData(appState.DB)

	log.Println("服务器启动在 http://localhost:8080")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器启动失败: %v\n", err)
	}
}
