package utils

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	DefualtSqlSafeLikeKeyword = 30
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

func SaveImages(c *gin.Context, file *multipart.FileHeader) (string, error) {
	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}
	//其实也可以直接用限定创建
	ext := filepath.Ext(file.Filename)
	//ext可以提取扩展名
	timestamp := time.Now().Unix()
	randomStr := uuid.New().String()[:8]
	newFileName := fmt.Sprintf("Cover_%d_%s%s", timestamp, randomStr, ext)

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
