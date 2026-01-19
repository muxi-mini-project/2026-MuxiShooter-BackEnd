package middleware

import (
	config "MuXi/2026-MuxiShooter-Backend/config"
	models "MuXi/2026-MuxiShooter-Backend/models"
	"net/http"
	"strings"
	"time"

	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"

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
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.Response{
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
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.Response{
				Code:    http.StatusUnauthorized, //401
				Message: "无效的token",
			})
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("user_id", claims["user_id"])
			c.Set("group", claims["group"])
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.Response{
				Code:    http.StatusUnauthorized, //401
				Message: "无效的token声明",
			})
			return
		}

		c.Next()
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")

		if userID == nil || !exists {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code:    http.StatusUnauthorized, //401
				Message: "请先登录",
			})
			c.Abort()
			//阻止后续中间件的执行
			return
		}

		c.Next()
		//继续处理后面的中间件
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		g, exists := c.Get("group")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.Response{
				Code:    http.StatusUnauthorized, //401
				Message: "token权限组参数缺失",
			})
			return
		}
		if group, ok := g.(string); !ok {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code:    http.StatusUnauthorized, //401
				Message: "权限组参数格式错误",
			})
			c.Abort()
			return
		} else if group != "admin" {
			c.JSON(http.StatusForbidden, models.Response{
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
