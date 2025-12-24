// @title MuXi Library Management System
// @version 1.0
// @description MuXi Library Management System
// @description 主要功能：用户登录注册（存在user和admin权限组），对图书的CRUD，可以借书和还书等等
// @description 其中增删改书籍权限仅限管理员账户，其他功能普通用户可用
// @termsOfService http://swagger.io/terms
// @contact.name FireChickenMP4
// @contact.email 13930176445@163.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey ApikeyAuth
// @in header
// @name Authorization
package main

import (
	"MuXi/Library/config"
	_ "MuXi/Library/docs"
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
	r.Static("/uploads", "./uploads")
	//gin的Static是Gin框架中用来提供静态文件服务的功能，就像在餐厅里设置一个自助区
	//让顾客可以自己取用饮料和小食，而不需要每次都找服务员点单。

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	//使用gzip传输

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
