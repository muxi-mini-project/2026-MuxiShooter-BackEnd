package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	models "MuXi/2026-MuxiShooter-Backend/models"
	utils "MuXi/2026-MuxiShooter-Backend/utils"
)

const (
	NumLimter                = 20
	DefaultHeadImagePath     = "static/DefaultHeadImg.jpeg"
	PasswordUpdatedInterval  = 30 * time.Minute
	UsernameUpdatedInterval  = 24 * time.Hour
	HeadImageUpdatedInterval = 24 * time.Hour
	PrefixHeadImg            = "HeadImg"
)

var (
	DB                       *gorm.DB
	JWTSecret                []byte
	ErrJWTWrongSigningMethod = errors.New("无效的签名算法")
	ErrJWTSecretGenerate     = errors.New("JWT密钥生成失败")
)

func ConnectDB() {
	log.Println("开始连接数据库...")
	dbUser := utils.GetEnv("DB_USER", "adminuser")
	dbPassword := utils.GetEnv("DB_PASSWORD", "")
	//我自己设置环境变量
	if len(dbPassword) == 0 {
		log.Fatal("数据库管理用户密码环境变量(DB_PASSWORD)为空,请配置")
	}
	dbHost := utils.GetEnv("DB_HOST", "")
	if dbHost == "" {
		log.Fatal("DB_HOST为空，请设置环境变量")
		return
	}
	dbPort := utils.GetEnv("DB_PORT", "3306")
	dbName := utils.GetEnv("DB_NAME", "mini")
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
		return
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

	err = DB.AutoMigrate(&models.Achievement{}, &models.User{}, &models.Skill{}, &models.Card{}, &models.Item{}, &models.UserAchievement{}, &models.UserCard{}, &models.UserItem{}, &models.UserSkill{})
	if err != nil {
		log.Fatal("数据迁移失败:", err)
	}
}

func InitAdmin(db *gorm.DB) {
	var err error
	admin := utils.GetEnv("ADMIN_USERNAME", "adminuser")

	adminPsw := utils.GetEnv("ADMIN_PASSWORD", "")

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
		Username:      admin,
		Password:      hashedPsw,
		Group:         "admin",
		HeadImagePath: DefaultHeadImagePath,
	}

	if err := db.Create(&adminUser).Error; err != nil {
		log.Fatal("初始化管理员失败:", err)
	}

	log.Printf("成功初始化管理员: %s / %s\n", admin, adminPsw)
}

func InitJWTSecret() {
	var err error
	secretStr := utils.GetEnv("JWT_SECRET", "")

	if len(secretStr) == 0 {
		log.Println("JWT密钥环境变量为空(JWT_SECRET),将随机生成")

		JWTSecret, err = utils.GenerateSercet(32)
		if err != nil {
			log.Fatal(ErrJWTSecretGenerate.Error() + ":" + err.Error())
		}
	} else {
		decoded, err := base64.StdEncoding.DecodeString(secretStr)
		if err == nil {
			if len(decoded) < 32 {
				log.Fatal("JWT密钥长度不足32字节")
			}
			JWTSecret = decoded
			log.Println("已使用JWT密钥环境变量(JWT_SECRET)")
		} else {
			log.Fatal("base64解码JWT密钥环境变量失败:" + err.Error())
		}
	}
}
