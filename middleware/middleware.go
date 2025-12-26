package middleware

import (
	"MuXi/Library/config"
	"MuXi/Library/models"
	"MuXi/Library/utils"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func InitSession(r *gin.Engine) {
	var err error
	sessionSecret := make([]byte, 32)
	sessionSecret = []byte(utils.GetEnv("SESSION_SECRET", ""))
	//这里要base64格式的
	if len(sessionSecret) == 0 {
		log.Println("session密钥环境变量为空(SESSION_SECRET),将随机生成")
		sessionSecret, err = utils.GenerateSessionSercet(32)
		if err != nil {
			log.Fatal(config.ErrSessionSecretGenerate.Error() + ":" + err.Error())
		}
	}

	var store sessions.Store
	store = initCookieStore(sessionSecret)

	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   gin.Mode() == gin.ReleaseMode,
		SameSite: 0,
	})

	r.Use(sessions.Sessions("LibSession", store))

	log.Println("Session 中间件初始化完成")
}

func initCookieStore(sessionSecret []byte) sessions.Store {
	log.Printf("使用Cookie存储Session")
	store := cookie.NewStore(sessionSecret)
	return store
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")

		if userID == nil {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code:    http.StatusUnauthorized, //401
				Message: "请先登录",
			})
			c.Abort()
			//阻止后续中间件的执行
			return
		}

		c.Set("user_id", userID)

		c.Next()
		//继续处理后面的中间件
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		g := session.Get("group")

		if group, ok := g.(string); !ok {
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    http.StatusBadRequest, //400
				Message: "权限组参数错误",
			})
			c.Abort()
			return
		} else if group != "admin" {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code:    http.StatusUnauthorized, //401
				Message: "请先登录",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
