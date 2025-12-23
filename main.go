package main

import (
	"MuXi/Library/config"
	_ "MuXi/Library/docs"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
)

func main() {
	config.ConnectDB()
	config.InitAdmin(config.DB)

	r := gin.Default()

}
