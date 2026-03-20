package middleware

import (
	config "MuXi/2026-MuxiShooter-Backend/config"
	"MuXi/2026-MuxiShooter-Backend/dto"
	"MuXi/2026-MuxiShooter-Backend/models"
	"net/http"
	"strings"
	"time"

	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

type JWTTokenParser interface {
	ParseToken(tokenStr string) (jwt.MapClaims, error)
}

type JWTUserRepository interface {
	FindByID(userID uint) (*models.User, bool, error)
}

func JWTAuth(tokenParser JWTTokenParser, userRepository JWTUserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tokenParser == nil || userRepository == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, dto.Response{
				Code:    http.StatusInternalServerError,
				Message: "鉴权组件未初始化",
			})
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized, //401
				Message: "请先登录",
			})
			return
		}

		authHeader = strings.TrimSpace(authHeader)
		var tokenStr string
		tokenStr = authHeader
		for {
			if len(tokenStr) < 6 || !strings.EqualFold(tokenStr[:6], "Bearer") {
				break
			}
			tokenStr = strings.TrimSpace(tokenStr[6:])
			tokenStr = strings.TrimSpace(strings.TrimPrefix(tokenStr, ":"))
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized,
				Message: "Authorization格式错误",
			})
			return
		}
		if strings.Count(tokenStr, ".") != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized,
				Message: "token格式错误，请传入登录接口返回的token",
			})
			return
		}

		claims, err := tokenParser.ParseToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized, //401
				Message: err.Error(),
			})
			return
		}

		userIDValue, uexists := claims["user_id"]
		groupValue, gexists := claims["group"]
		tokenVersionValue, texists := claims["token_version"]

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

		user, existed, err := userRepository.FindByID(userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, dto.Response{
				Code:    http.StatusInternalServerError, //500
				Message: "查询数据库失败：" + err.Error(),
			})
			return
		}
		if !existed || user == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.Response{
				Code:    http.StatusForbidden, //403
				Message: "用户不存在",
			})
			return
		}

		var tokenVersion uint64
		switch v := tokenVersionValue.(type) {
		case float64:
			if v <= 0 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
					Code:    http.StatusUnauthorized, //401
					Message: "无效的TokenVersion",
				})
				return
			}
			tokenVersion = uint64(v)
		case int64:
			if v <= 0 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
					Code:    http.StatusUnauthorized, //401
					Message: "无效的TokenVersion",
				})
				return
			}
			tokenVersion = uint64(v)
		case int:
			if v <= 0 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
					Code:    http.StatusUnauthorized, //401
					Message: "无效的TokenVersion",
				})
				return
			}
			tokenVersion = uint64(v)
		case uint:
			if v == 0 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
					Code:    http.StatusUnauthorized, //401
					Message: "无效的TokenVersion",
				})
				return
			}
			tokenVersion = uint64(v)
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized, //401
				Message: "TokenVersion格式错误",
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

func PaginationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var pagination models.Pagination

		if err := c.ShouldBindQuery(&pagination); err != nil {
			pagination = models.Pagination{
				Page:     config.DefaultPage,
				PageSize: config.DefaultPageSize,
			}
		}

		if pagination.Page <= 0 {
			pagination.Page = config.DefaultPage
		}
		if pagination.PageSize <= 0 {
			pagination.PageSize = config.DefaultPageSize
		}
		if pagination.PageSize > config.MaxPageSize {
			pagination.PageSize = config.MaxPageSize
		}

		pagination.Limit = pagination.PageSize
		pagination.Offset = (pagination.Page - 1) * pagination.PageSize

		c.Set("pagination", pagination)

		c.Next()
	}
}

func GetPagination(c *gin.Context) models.Pagination {
	if val, exists := c.Get("pagination"); exists {
		if p, ok := val.(models.Pagination); ok {
			return p
		}
	}

	//否则返回一个安全的默认值
	return models.Pagination{
		Page:     config.DefaultPage,
		PageSize: config.DefaultPageSize,
		Limit:    config.DefaultPageSize,
		Offset:   0,
	}
}
