package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"MuXi/Library/models"
	"MuXi/Library/utils"
)

const (
	DefualtCoverPath = "uploads/default.png"
	DefualtSummary   = "这里空空如也"
)

var (
	DB                        *gorm.DB
	ErrBookNotFound           = errors.New("图书不存在")
	ErrNoStock                = errors.New("图书库存不足")
	ErrBorrowedRecordNotFound = errors.New("借书记录查询失败")
	ErrBookBorrowed           = errors.New("图书在借")
	ErrDeleteBook             = errors.New("图书删除失败")
	ErrDeleteCover            = errors.New("封面删除失败")
)

func getEnv(key, def string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return def
}

func ConnectDB() {
	dbUser := getEnv("DB_USER", "adminuser")
	dbPassword := getEnv("DB_PASSWORD", "")
	//我自己设置环境变量
	dbHost := getEnv("DB_HOST", "47.105.123.226")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "Lib")
	dsnRoot := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=true&loc=Local",
		dbUser, dbPassword, dbHost, dbPort)
	var err error

	maxRetries := 15
	for i := range maxRetries {
		DB, err = gorm.Open(mysql.Open(dsnRoot), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("启动中(%d/%d),Error:%v", i+1, maxRetries, err)
		time.Sleep(1e9)
	}

	if err != nil {
		log.Fatal("连接MySQL失败:", err)
	}
	createDb := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4;", dbName)
	err = DB.Exec(createDb).Error
	if err != nil {
		log.Fatal("创建数据库失败:", err)
	}
	//确保存在数据库，IF EXISTS 判断
	dsnDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)
	DB, err = gorm.Open(mysql.Open(dsnDB), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	log.Println("数据库连接成功")

	err = DB.AutoMigrate(&models.Book{}, &models.User{}, &models.BorrowRecord{})
	if err != nil {
		log.Fatal("数据迁移失败:", err)
	}
}

func InitAdmin(db *gorm.DB) {
	var err error
	admin := getEnv("ADMIN_USERNAME", "adminuser")

	adminPsw := getEnv("ADMIN_PASSWORD", "")

	var count int64
	db.Model(&models.User{}).Where("username = ?", admin).Count(&count)

	if count > 0 {
		log.Println("管理员账户已存在，跳过初始化")
		return
	}

	hashedPsw, err := utils.Hashtool(adminPsw)
	if err != nil {
		log.Fatal("管理员密码哈希失败:", err)
	}

	adminUser := models.User{
		Username: admin,
		Password: hashedPsw,
		Group:    "admin",
	}

	if err := db.Create(&adminUser).Error; err != nil {
		log.Fatal("初始化管理员失败:", err)
	}

	log.Printf("成功初始化管理员: %s / %s\n", admin, adminPsw)
}
