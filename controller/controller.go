package controller

import (
	config "MuXi/2026-MuxiShooter-Backend/config"
	"MuXi/2026-MuxiShooter-Backend/dto"
	models "MuXi/2026-MuxiShooter-Backend/models"
	utils "MuXi/2026-MuxiShooter-Backend/utils"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

// @Summary		用户注册
// @Description	注册用户
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		models.RegisterRequest						true	"注册请求"
// @Success		200		{object}	dto.Response{data=dto.AuthData}	"注册成功"
// @Failure		400		{object}	dto.Response								"请求参数错误"
// @Failure		409		{object}	dto.Response								"用户已存在"
// @Failure		500		{object}	dto.Response								"服务器错误"
// @Router			/api/auth/register [post]
func Register(c *gin.Context) {
	var req dto.RegisterRequest
	var err error

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest, //400
			Message: "请求参数错误",
		})
		return
	}

	var searchedUser models.User
	err = config.DB.Where("username = ?", req.UserName).First(&searchedUser).Error
	//这里不用first的话就要用users切片，然后Find(&users)
	//我们只需要自己确保只有一个就ok
	if err == nil {
		c.JSON(http.StatusConflict, dto.Response{
			Code:    http.StatusConflict, //409
			Message: "用户已存在",
		})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "查询数据库失败：" + err.Error(),
		})
		return
	}
	//notfound就可以注册了

	hashedPsw, err := utils.Hashtool(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "注册密码哈希失败：" + err.Error(),
		})
		return
	}

	newUser := models.User{
		Username: req.UserName,
		Password: hashedPsw,
		Group:    "user",
	}

	if err = config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "注册用户失败：" + err.Error(),
		})
		return
	}

	//Token过期时间,24h
	expirationTime := time.Now().Add(24 * time.Hour)

	//创建claims
	claims := jwt.MapClaims{
		"user_id": newUser.ID,
		"group":   newUser.Group,
		"exp":     expirationTime.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString(config.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "生成Token失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK, //200 ok
		Message: "注册用户成功",
		Data: dto.AuthData{
			User: dto.CommonUserData{
				UserID:        newUser.ID,
				Username:      newUser.Username,
				Group:         newUser.Group,
				HeadImagePath: newUser.HeadImagePath,
				StrengthCoin:  newUser.StrengthCoin,
				SelectCoin:    newUser.SelectCoin,
			},
			Token:     tokenStr,
			ExpiresAt: expirationTime.Unix(),
		},
	})
}

// @Summary		用户登录
// @Description	用户登录
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		models.LoginRequest						true	"注册请求"
// @Success		200		{object}	dto.Response{data=dto.CommonUserData}	"登录成功"
// @Failure		400		{object}	dto.Response							"请求参数错误"
// @Failure		403		{object}	dto.Response							"认证失败"
// @Failure		500		{object}	dto.Response							"服务器错误"
// @Router			/api/auth/login [post]
func Login(c *gin.Context) {
	var req dto.LoginRequest
	var err error

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest, //400
			Message: "请求参数错误",
		})
		return
	}

	var user models.User
	err = config.DB.Where("username = ?", req.UserName).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusForbidden, dto.Response{
			Code:    http.StatusForbidden, //403
			Message: "用户不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "查询数据库失败：" + err.Error(),
		})
		return
	}

	err = utils.ComparePassword(user.Password, req.Password)
	if err != nil {
		c.JSON(http.StatusForbidden, dto.Response{
			Code:    http.StatusForbidden, //403
			Message: "密码错误",
		})
		return
	}

	//Token过期时间,24h
	expirationTime := time.Now().Add(24 * time.Hour)

	//创建claims
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"group":   user.Group,
		"exp":     expirationTime.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString(config.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "生成Token失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "登录成功",
		Data: dto.AuthData{
			User: dto.CommonUserData{
				UserID:        user.ID,
				Username:      user.Username,
				Group:         user.Group,
				HeadImagePath: user.HeadImagePath,
				StrengthCoin:  user.StrengthCoin,
				SelectCoin:    user.SelectCoin,
			},
			Token:     tokenStr,
			ExpiresAt: expirationTime.Unix(),
		},
	})
}
