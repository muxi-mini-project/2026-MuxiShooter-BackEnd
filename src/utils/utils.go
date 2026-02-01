package utils

import (
	"MuXi/2026-MuxiShooter-Backend/models"
	"crypto/rand"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	DefualtSqlSafeLikeKeyword = 30
	TokenExpirationTime       = 24 * time.Hour
	uploadDir                 = "uploads"
)

var (
	ErrTokenGenerate = errors.New("Token生成失败:")
	ErrTokenExpired  = errors.New("Token已失效，请重新登陆")
	ErrUserNotFound  = errors.New("用户不存在")
)

func Hashtool(key string) (string, error) {
	//工具的话还是返回错误比较好
	var err error
	hashedkey, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	//这里用现成的加密包bcrypt
	if err != nil {
		return "", err
	}
	return string(hashedkey), nil
}

func ComparePassword(dbPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

func SaveImages(c *gin.Context, file *multipart.FileHeader, prefix string) (string, error) {
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}
	//其实也可以直接用限定创建
	ext := filepath.Ext(file.Filename)
	//ext可以提取扩展名
	timestamp := time.Now().Unix()
	randomStr := uuid.New().String()[:8]
	newFileName := fmt.Sprintf("%s_%d_%s%s", prefix, timestamp, randomStr, ext)

	dst := filepath.Join(uploadDir, newFileName)
	//连接路径

	if err := c.SaveUploadedFile(file, dst); err != nil {
		return "", err
	}
	//没想到gin本身就有保存文件啊，好好gin
	return filepath.ToSlash(dst), nil
	//统一路径分隔符,会根据系统选择\还是/
}

func RemoveFile(filePath string) error {
	if filePath != "" {
		err := os.Remove(filePath)
		return err
	}
	return nil
	//没有也不用管，反正本身就没
}

func SqlSafeLikeKeyword(input string) string {
	//默认最多三十字
	if len(input) > DefualtSqlSafeLikeKeyword {
		input = input[:DefualtSqlSafeLikeKeyword]
	}

	//转义%和_ 防止轰炸。。。
	//这里用到strings包
	input = strings.ReplaceAll(input, "%", "\\%")
	input = strings.ReplaceAll(input, "_", "\\_")

	//禁止单独通配符
	if input == "%" || input == "_" || input == "" {
		return ""
		//直接无效
	}

	return input
}

func GetEnv(key, def string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return def
}

func GenerateSercet(keyLength int) ([]byte, error) {
	key := make([]byte, keyLength)
	_, err := rand.Read(key)

	return key, err
}

func GenerateToken(user models.User, jwtSecret []byte) (tokenStr string, expirationTime time.Time, err error) {
	//Token过期时间,24h
	expirationTime = time.Now().Add(TokenExpirationTime)

	//创建claims
	//版号按照对应的来
	claims := jwt.MapClaims{
		"user_id":       user.ID,
		"group":         user.Group,
		"token_version": user.TokenVersion,
		"exp":           expirationTime.Unix(),
		"iat":           time.Now().Unix(),
		"jti":           uuid.New().String(), // JWT ID用于可能存在的单token吊销
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err = token.SignedString(jwtSecret)
	if err != nil {
		err = fmt.Errorf("%w%w", ErrTokenGenerate, err)
		return "", expirationTime, err
	} else {
		return tokenStr, expirationTime, nil
	}
}

func RefreshToken(userID uint, db *gorm.DB) error {
	//只是刷新，不负责生成token，也就是只是递增token版号
	//注意：请确保id有效，否者没法增加对应的token版号
	result := db.Model(&models.User{}).
		Set("gorm:query_option", "FOR UPDATE").
		Where("id = ?", userID).
		UpdateColumn("token_version", gorm.Expr("token_version + 1"))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		//检查是否真的更新了记录
		return ErrUserNotFound
	}

	return nil
}
