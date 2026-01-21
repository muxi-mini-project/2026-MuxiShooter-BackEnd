package middleware

import (
	config "MuXi/2026-MuxiShooter-Backend/config"
	"MuXi/2026-MuxiShooter-Backend/dto"
	"MuXi/2026-MuxiShooter-Backend/models"
	"errors"
	"net/http"
	"strings"
	"time"

	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized, //401
				Message: "请先登录",
			})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, config.ErrJWTWrongSigningMethod
			}
			secret := config.JWTSecret
			return secret, nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized, //401
				Message: err.Error(),
			})
			return
		}
		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized, //401
				Message: "无效的token",
			})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userIDValue, uexists := claims["user_id"]
			groupValue, gexists := claims["group"]
			tokenVersion, texists := claims["token_version"]

			if !texists {
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
					Code:    http.StatusUnauthorized, //401
					Message: "缺少token版本号",
				})
				return
			}

			if !uexists {
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
					Code:    http.StatusUnauthorized, //401
					Message: "token中缺少用户id信息",
				})
				return
			}
			var userID uint
			switch v := userIDValue.(type) {
			case float64:
				if v <= 0 {
					c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
						Code:    http.StatusUnauthorized, //401
						Message: "无效的用户ID",
					})
					return
				}
				userID = uint(v)
			case int:
				if v <= 0 {
					c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
						Code:    http.StatusUnauthorized, //401
						Message: "无效的用户ID",
					})
					return
				}
				userID = uint(v)
			default:
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
					Code:    http.StatusUnauthorized, //401
					Message: "用户ID格式错误",
				})
				return
			}

			var user models.User
			err = config.DB.Where("id = ?", userID).First(&user).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(http.StatusForbidden, dto.Response{
					Code:    http.StatusForbidden, //403
					Message: "用户不存在",
				})
				return
			}
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, dto.Response{
					Code:    http.StatusInternalServerError, //500
					Message: "查询数据库失败：" + err.Error(),
				})
				return
			}
			if tokenVersion != user.TokenVersion {
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
					Code:    http.StatusUnauthorized, //401
					Message: "token版本号错误",
				})
				return
			}
			c.Set("user_id", userID)
			if gexists {
				if groupStr, ok := groupValue.(string); ok && groupStr != "" {
					c.Set("group", groupStr)
				} else {
					c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
						Code:    http.StatusUnauthorized, //401
						Message: "用户权限组错误",
					})
					return
				}
			}
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized, //401
				Message: "无效的token声明",
			})
			return
		}

		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		g, exists := c.Get("group")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized, //401
				Message: "token权限组参数缺失",
			})
			return
		}
		if group, ok := g.(string); !ok {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized, //401
				Message: "权限组参数格式错误",
			})
			c.Abort()
			return
		} else if group != "admin" {
			c.JSON(http.StatusForbidden, dto.Response{
				Code:    http.StatusForbidden, //403
				Message: "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func Limiter() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: 1 * time.Second,
		Limit:  config.NumLimter, //20
	}
	store := memory.NewStore()

	middleware := mgin.NewMiddleware(limiter.New(store, rate))
	return middleware
}
