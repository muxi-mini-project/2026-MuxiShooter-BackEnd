package controller

import (
	config "MuXi/2026-MuxiShooter-Backend/config"
	models "MuXi/2026-MuxiShooter-Backend/models"
	utils "MuXi/2026-MuxiShooter-Backend/utils"
	"errors"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary		用户注册
// @Description	注册用户
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		models.RegisterRequest						true	"注册请求"
// @Success		200		{object}	models.Response{data=models.RegisterData}	"注册成功"
// @Failure		400		{object}	models.Response								"请求参数错误"
// @Failure		409		{object}	models.Response								"用户已存在"
// @Failure		500		{object}	models.Response								"服务器错误"
// @Router			/api/auth/register [post]
func Register(c *gin.Context) {
	var req models.RegisterRequest
	var err error

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    http.StatusBadRequest, //400
			Message: "请求参数错误",
		})
		return
	}

	var searchedUser models.User
	err = config.DB.Where("username = ?", req.Name).First(&searchedUser).Error
	//这里不用first的话就要用users切片，然后Find(&users)
	//我们只需要自己确保只有一个就ok
	if err == nil {
		c.JSON(http.StatusConflict, models.Response{
			Code:    http.StatusConflict, //409
			Message: "用户已存在",
		})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "查询数据库失败：" + err.Error(),
		})
		return
	}
	//notfound就可以注册了

	hashedPsw, err := utils.Hashtool(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError,
			Message: "注册密码哈希失败：" + err.Error(),
		})
		return
	}

	newUser := models.User{
		Name:     req.Name,
		Password: hashedPsw,
		Group:    "user",
	}

	if err = config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError,
			Message: "注册用户失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    http.StatusOK, //200 ok
		Message: "注册用户成功",
		Data: gin.H{
			"username": newUser.Name,
			"user_id":  newUser.ID,
		},
	})
}

// @Summary		用户登录
// @Description	用户登录
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		models.LoginRequest						true	"注册请求"
// @Success		200		{object}	models.Response{data=models.LoginData}	"登录成功"
// @Failure		400		{object}	models.Response							"请求参数错误"
// @Failure		403		{object}	models.Response							"认证失败"
// @Failure		500		{object}	models.Response							"服务器错误"
// @Router			/api/auth/login [post]
func Login(c *gin.Context) {
	var req models.LoginRequest
	var err error

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    http.StatusBadRequest, //400
			Message: "请求参数错误",
		})
		return
	}

	var user models.User
	err = config.DB.Where("username = ?", req.Name).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    http.StatusForbidden, //403
			Message: "用户不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "查询数据库失败：" + err.Error(),
		})
		return
	}

	err = utils.ComparePassword(user.Password, req.Password)
	if err != nil {
		c.JSON(http.StatusForbidden, models.Response{
			Code:    http.StatusForbidden, //403
			Message: "密码错误",
		})
		return
	}

	//接下来获取所属权限组
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("group", user.Group)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "鉴权组件错误：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    http.StatusOK,
		Message: "登录成功",
		Data: gin.H{
			"user_id": user.ID,
			"group":   user.Group,
		},
	})
}
