package controller

import (
	config "MuXi/2026-MuxiShooter-Backend/config"
	"MuXi/2026-MuxiShooter-Backend/dto"
	"MuXi/2026-MuxiShooter-Backend/middleware"
	models "MuXi/2026-MuxiShooter-Backend/models"
	utils "MuXi/2026-MuxiShooter-Backend/utils"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary		用户注册
// @Description	注册用户
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.RegisterRequest				true	"注册请求"
// @Success		200		{object}	dto.Response{data=dto.AuthData}	"注册成功"
// @Failure		400		{object}	dto.Response					"请求参数错误"
// @Failure		409		{object}	dto.Response					"用户已存在"
// @Failure		500		{object}	dto.Response					"服务器错误"
// @Router			/api/auth/register [post]
func Register(c *gin.Context) {
	var req dto.RegisterRequest
	var err error

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest, //400
			Message: "请求参数错误:" + err.Error(),
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
		Username:      req.UserName,
		Password:      hashedPsw,
		Group:         "user",
		HeadImagePath: config.DefaultHeadImagePath,
	}

	if err = config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "注册用户失败：" + err.Error(),
		})
		return
	}

	//Token过期时间,24h
	token, expirationTime, err := utils.GenerateToken(newUser, config.JWTSecret)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: err.Error(),
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
			Token:     token,
			ExpiresAt: expirationTime.Unix(),
		},
	})
}

// @Summary		用户登录
// @Description	用户登录
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.LoginRequest				true	"注册请求"
// @Success		200		{object}	dto.Response{data=dto.AuthData}	"登录成功"
// @Failure		400		{object}	dto.Response					"请求参数错误"
// @Failure		403		{object}	dto.Response					"认证失败"
// @Failure		500		{object}	dto.Response					"服务器错误"
// @Router			/api/auth/login [post]
func Login(c *gin.Context) {
	var req dto.LoginRequest
	var err error

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest, //400
			Message: "请求参数错误:" + err.Error(),
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
	token, expirationTime, err := utils.GenerateToken(user, config.JWTSecret)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: err.Error(),
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
			Token:     token,
			ExpiresAt: expirationTime.Unix(),
		},
	})
}

// @Summary		用户登出（token版号增加一）
// @Description	用户登出（token版号增加一）
// @Tags			profile-operation
// @Produce		json
// @Success		200	{object}	dto.Response	"登出成功"
// @Failure		401	{object}	dto.Response	"用户不存在"
// @Failure		500	{object}	dto.Response	"服务器错误"
// @Router			/api/profile/operation/logout [get]
func Logout(c *gin.Context) {
	var err error
	userId, uexists := c.Get("user_id")
	if !uexists {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized, //401
			Message: "解析后token中缺少用户信息",
		})
		return
	}
	userID, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized, //401
			Message: "用户不存在",
		})
		return
	}
	err = utils.RefreshToken(userID, config.DB)
	if err != nil {
		if err == utils.ErrUserNotFound {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized, //401
				Message: "用户不存在",
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, dto.Response{
				Code:    http.StatusInternalServerError, //500
				Message: fmt.Sprintf("服务器错误: %v, 请重试", err),
			})
			return
		}
	}
	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK, //200
		Message: "登出成功",
	})
}

// @Summary		修改用户密码
// @Description	修改用户密码(修改完前端请删掉token并跳转登录页面)
// @Tags			profile-update
// @Accept			json
// @Produce		json
// @Param			request	body		dto.UpdatePasswordRequest	true	"修改密码请求"
// @Success		200		{object}	dto.Response				"修改密码成功"
// @Failure		400		{object}	dto.Response				"请求参数错误"
// @Failure		401		{object}	dto.Response				"登录状态异常"
// @Failure		403		{object}	dto.Response				"认证失败"
// @Failure		404		{object}	dto.Response				"用户不存在"
// @Failure		500		{object}	dto.Response				"服务器错误"
// @Router			/api/profile/update/password [put]
func UpdatePassword(c *gin.Context) {
	var req dto.UpdatePasswordRequest
	var err error

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest, //400
			Message: "请求参数错误:" + err.Error(),
		})
		return
	}
	if req.NewPassword == req.OldPassword {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest, //400
			Message: "所给新旧密码不能相同",
		})
		return
	}

	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized, //401
			Message: "解析后token中缺少用户信息",
		})
		return
	}
	userID, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized, //401
			Message: "解析后token中缺少用户信息",
		})
		return
	}

	var user models.User

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := tx.Set("gorm:query_option", "FOR UPDATE").First(&user, userId)

	err = result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.JSON(http.StatusNotFound, dto.Response{
				Code:    http.StatusNotFound, //404
				Message: "用户不存在",
			})
		} else {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, dto.Response{
				Code:    http.StatusInternalServerError, //500
				Message: "数据库查询失败：" + err.Error(),
			})
		}
		return
	}

	ok = func() bool {
		if user.PasswordUpdatedAt == nil {
			return true
		} else {
			return time.Since(*user.PasswordUpdatedAt) > config.PasswordUpdatedInterval
		}
	}()

	if !ok {
		tx.Rollback()
		c.JSON(http.StatusForbidden, dto.Response{
			Code:    http.StatusForbidden, //403
			Message: "修改密码间隔过短",
		})
		return
	}

	err = utils.ComparePassword(user.Password, req.OldPassword)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusForbidden, dto.Response{
			Code:    http.StatusForbidden, //403
			Message: "旧密码错误",
		})
		return
	}

	hashedPassword, err := utils.Hashtool(req.NewPassword)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "哈希新密码失败：" + err.Error(),
		})
		return
	}
	now := time.Now()
	updates := make(map[string]interface{})
	updates["password"] = hashedPassword
	updates["password_updated_at"] = &now
	if err = tx.Model(&user).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "修改密码失败：" + err.Error(),
		})
		return
	}

	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "提交Update请求失败：" + err.Error(),
		})
		return
	}

	utils.RefreshToken(userID, config.DB)

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "修改密码成功，已退出登录，请重新登陆",
	})
}

// @Summary		修改用户名
// @Description	修改用户名
// @Tags			profile-update
// @Accept			json
// @Produce		json
// @Param			request	body		dto.UpdateUsernameRequest	true	"修改用户名请求"
// @Success		200		{object}	dto.Response				"修改用户名成功"
// @Failure		400		{object}	dto.Response				"请求参数错误"
// @Failure		401		{object}	dto.Response				"登录状态异常"
// @Failure		403		{object}	dto.Response				"认证失败"
// @Failure		404		{object}	dto.Response				"用户不存在"
// @Failure		500		{object}	dto.Response				"服务器错误"
// @Router			/api/profile/update/username [put]
func UpdateUsername(c *gin.Context) {
	var req dto.UpdateUsernameRequest
	var err error

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest, //400
			Message: "请求参数错误:" + err.Error(),
		})
		return
	}

	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized, //401
			Message: "解析后token中缺少用户信息",
		})
		return
	}

	var user models.User

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := tx.Set("gorm:query_option", "FOR UPDATE").First(&user, userId)

	err = result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.JSON(http.StatusNotFound, dto.Response{
				Code:    http.StatusNotFound, //404
				Message: "用户不存在",
			})
		} else {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, dto.Response{
				Code:    http.StatusInternalServerError, //500
				Message: "数据库查询失败：" + err.Error(),
			})
		}
		return
	}

	ok := func() bool {
		if user.UsernameUpdatedAt == nil {
			return true
		} else {
			return time.Since(*user.UsernameUpdatedAt) > config.UsernameUpdatedInterval
		}
	}()

	if !ok {
		tx.Rollback()
		c.JSON(http.StatusForbidden, dto.Response{
			Code:    http.StatusForbidden, //403
			Message: "修改用户名间隔过短",
		})
		return
	}

	now := time.Now()
	updates := make(map[string]interface{})
	updates["username"] = req.NewUsername
	updates["username_updated_at"] = &now
	if err = tx.Model(&user).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "修改用户名失败：" + err.Error(),
		})
		return
	}

	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "提交Update请求失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "修改用户名成功",
	})
}

// @Summary		修改用户头像
// @Description	修改用户头像
// @Tags			profile-update
// @Accept			multipart/form-data
// @Produce		json
// @Param			new_head_image	formData	file			true	"新头像"
// @Success		200				{object}	dto.Response	"登录成功"
// @Failure		400				{object}	dto.Response	"头像为空"
// @Failure		401				{object}	dto.Response	"登录状态异常"
// @Failure		403				{object}	dto.Response	"认证失败"
// @Failure		404				{object}	dto.Response	"用户不存在"
// @Failure		500				{object}	dto.Response	"服务器错误"
// @Router			/api/profile/update/headimage [put]
func UpdateHeadImage(c *gin.Context) {
	var req dto.UpdateHeadImageRequest
	var err error

	if err = c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest, //400
			Message: "请求参数错误:" + err.Error(),
		})
		return
	}

	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized, //401
			Message: "解析后token中缺少用户信息",
		})
		return
	}

	if req.NewHeadImage == nil || req.NewHeadImage.Size == 0 {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest, //400
			Message: "头像为空",
		})
		return
	}

	var user models.User

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := tx.Set("gorm:query_option", "FOR UPDATE").First(&user, userId)

	err = result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.JSON(http.StatusNotFound, dto.Response{
				Code:    http.StatusNotFound, //404
				Message: "用户不存在",
			})
		} else {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, dto.Response{
				Code:    http.StatusInternalServerError, //500
				Message: "数据库查询失败：" + err.Error(),
			})
		}
		return
	}

	ok := func() bool {
		if user.HeadImageUpdatedAt == nil {
			return true
		} else {
			return time.Since(*user.HeadImageUpdatedAt) > config.HeadImageUpdatedInterval
		}
	}()

	if !ok {
		tx.Rollback()
		c.JSON(http.StatusForbidden, dto.Response{
			Code:    http.StatusForbidden, //403
			Message: "修改头像间隔过短",
		})
		return
	}

	log.Printf("用户:%s(id:%d)上传头像,Size:%d", user.Username, user.ID, req.NewHeadImage.Size)
	savePath, err := utils.SaveImages(c, req.NewHeadImage, config.PrefixHeadImg)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "图片保存失败：" + err.Error(),
		})
		return
	}
	now := time.Now()
	updates := make(map[string]interface{})
	updates["head_image_path"] = savePath
	updates["head_image_updated_at"] = &now
	if err = tx.Model(&user).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "修改头像失败：" + err.Error(),
		})
		return
	}

	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "提交Update请求失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "修改头像成功",
	})
}

// @Summary		获取用户列表
// @Description	注: 管理员可用，查询结果是多重模糊搜索叠加的效果
// @Description 以及页码不输入或不合规范自动为第一页，每页多少不输入默认20，最多100
// @Description 如果查询结果不存在则返回切片为空
// @Description 用了id查询的话就一定只是一个确定的，而不是模糊搜索，其他参数就没用了（分页也是）
// @Tags	admin-get
// @Security		ApiKeyAuth
// @Produce		json
// @Param			id	query		int								false	"用户id"
// @Param			username	query		string								false	"用户名"
// @Param			group	query		string								false	"权限组(user/admin)"
// @Param			page	query		int								false	"页码，默认1"
// @Param			page_size	query		int								false	"每页多少，默认20，最大100"
// @Success		200		{object}	dto.Response{data=dto.PaginatedData}	"查询成功"
// @Failure		500		{object}	dto.Response						"数据库查询失败"
// @Router			/api/admin/get/getusers [get]
func GetUsers(c *gin.Context) {
	var err error
	pagination := middleware.GetPagination(c)

	id := c.Query("id")
	username := c.Query("username")
	group := c.Query("group")

	var users []models.User
	var total int64

	if id != "" {
		var user models.User
		pagination = models.Pagination{
			Page:     config.DefaultPage,
			PageSize: config.DefaultPageSize,
			Limit:    config.DefaultPageSize,
			Offset:   0,
		}
		err = config.DB.First(&user, id).Error
		if err != nil {

		}
	} else {
		//没id就再查别的，模糊搜索
	}

}
