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

const (
	initialAdminUserID uint = 1
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
	oldHeadImagePath := user.HeadImagePath
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
		if removeErr := utils.RemoveFile(savePath); removeErr != nil {
			log.Printf("回滚时删除新头像失败(user_id:%d,path:%s): %v", user.ID, savePath, removeErr)
		}
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "修改头像失败：" + err.Error(),
		})
		return
	}

	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		if removeErr := utils.RemoveFile(savePath); removeErr != nil {
			log.Printf("提交失败时删除新头像失败(user_id:%d,path:%s): %v", user.ID, savePath, removeErr)
		}
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError, //500
			Message: "提交Update请求失败：" + err.Error(),
		})
		return
	}

	if oldHeadImagePath != "" && oldHeadImagePath != savePath && oldHeadImagePath != config.DefaultHeadImagePath {
		if removeErr := utils.RemoveFile(oldHeadImagePath); removeErr != nil {
			log.Printf("删除旧头像失败(user_id:%d,path:%s): %v", user.ID, oldHeadImagePath, removeErr)
		}
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "修改头像成功",
	})
}

// @Summary		获取用户列表
// @Description	注: 管理员可用，查询结果是多重模糊搜索叠加的效果
// @Description	以及页码不输入或不合规范自动为第一页，每页多少不输入默认20，最多100
// @Description	如果查询结果不存在则返回切片为空
// @Description	用了id查询的话就一定只是一个确定的，而不是模糊搜索，其他参数就没用了（分页也是）
// @Tags			admin-get
// @Produce		json
// @Param			user_id		query		int										false	"用户id"
// @Param			username	query		string									false	"用户名"
// @Param			group		query		string									false	"权限组(user/admin)"
// @Param			page		query		int										false	"页码，默认1"
// @Param			page_size	query		int										false	"每页多少，默认20，最大100"
// @Success		200			{object}	dto.Response{data=dto.PaginatedData}	"查询成功"
// @Failure		401			{object}	dto.Response							"登录状态异常"
// @Failure		500			{object}	dto.Response							"数据库查询失败"
// @Router			/api/admin/get/getusers [get]
func GetUsers(c *gin.Context) {
	var err error
	pagination := middleware.GetPagination(c)

	id := c.Query("user_id")
	username := utils.SqlSafeLikeKeyword(c.Query("username"))
	group := utils.SqlSafeLikeKeyword(c.Query("group"))

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
		result := config.DB.First(&user, id)
		err = result.Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				total = 0
			} else {
				c.JSON(http.StatusInternalServerError, dto.Response{
					Code:    http.StatusInternalServerError, //500
					Message: "数据库查询失败：" + err.Error(),
				})
				return
			}
		} else {
			total = 1
		}
		if total != 0 {
			users = append(users, user)
		}
	} else {
		result := config.DB.Model(&models.User{})
		//没id就再查别的，模糊搜索
		if username != "" {
			result = result.Where("username LIKE ?", "%"+username+"%")
		}
		if group != "" {
			result = result.Where("group LIKE ?", "%"+group+"%")
		}
		result.Count(&total)
		result.Limit(pagination.Limit).Offset(pagination.Offset).Find(&users)
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "查询成功",
		Data: dto.PaginatedData{
			List:     users,
			Total:    total,
			Page:     pagination.Page,
			PageSize: pagination.PageSize,
		},
	})
}

// @Summary		管理员删除用户
// @Description	管理员删除用户。ID=1(初始化管理员)不可删除；其他管理员仅可删除普通用户
// @Tags			admin-operation
// @Accept			json
// @Produce		json
// @Param			request	body		dto.AdminDeleteUserRequest	true	"删除用户请求"
// @Success		200		{object}	dto.Response				"删除成功"
// @Failure		400		{object}	dto.Response				"请求参数错误"
// @Failure		401		{object}	dto.Response				"登录状态异常"
// @Failure		403		{object}	dto.Response				"权限不足"
// @Failure		404		{object}	dto.Response				"用户不存在"
// @Failure		500		{object}	dto.Response				"数据库错误"
// @Router			/api/admin/operation/deleteuser [delete]
func DeleteUserByAdmin(c *gin.Context) {
	var req dto.AdminDeleteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误:" + err.Error(),
		})
		return
	}

	operatorID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: "解析后token中缺少用户信息",
		})
		return
	}
	adminID, ok := operatorID.(uint)
	if !ok || adminID == 0 {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: "用户信息格式错误",
		})
		return
	}

	if req.UserID == initialAdminUserID {
		c.JSON(http.StatusForbidden, dto.Response{
			Code:    http.StatusForbidden,
			Message: "初始化管理员不可删除",
		})
		return
	}

	var targetUser models.User
	if err := config.DB.First(&targetUser, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{
				Code:    http.StatusNotFound,
				Message: "目标用户不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "数据库查询失败：" + err.Error(),
		})
		return
	}

	if adminID != initialAdminUserID && targetUser.Group != "user" {
		c.JSON(http.StatusForbidden, dto.Response{
			Code:    http.StatusForbidden,
			Message: "仅初始化管理员可删除管理员账户",
		})
		return
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "开启事务失败：" + tx.Error.Error(),
		})
		return
	}

	if err := tx.Delete(&targetUser).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "删除用户失败：" + err.Error(),
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "提交删除请求失败：" + err.Error(),
		})
		return
	}

	if targetUser.HeadImagePath != "" && targetUser.HeadImagePath != config.DefaultHeadImagePath {
		if removeErr := utils.RemoveFile(targetUser.HeadImagePath); removeErr != nil {
			log.Printf("删除用户头像文件失败(user_id:%d,path:%s): %v", targetUser.ID, targetUser.HeadImagePath, removeErr)
		}
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "删除用户成功",
	})
}

// @Summary		管理员修改用户权限组
// @Description	ID=1(初始化管理员)权限组不可修改；仅ID=1可修改其他用户权限组
// @Tags			admin-update
// @Accept			json
// @Produce		json
// @Param			request	body		dto.AdminUpdateUserGroupRequest	true	"修改权限组请求"
// @Success		200		{object}	dto.Response					"修改成功"
// @Failure		400		{object}	dto.Response					"请求参数错误"
// @Failure		401		{object}	dto.Response					"登录状态异常"
// @Failure		403		{object}	dto.Response					"权限不足"
// @Failure		404		{object}	dto.Response					"用户不存在"
// @Failure		500		{object}	dto.Response					"数据库错误"
// @Router			/api/admin/update/usergroup [put]
func UpdateUserGroupByAdmin(c *gin.Context) {
	var req dto.AdminUpdateUserGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误:" + err.Error(),
		})
		return
	}

	operatorID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: "解析后token中缺少用户信息",
		})
		return
	}
	adminID, ok := operatorID.(uint)
	if !ok || adminID == 0 {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: "用户信息格式错误",
		})
		return
	}

	if req.UserID == initialAdminUserID {
		c.JSON(http.StatusForbidden, dto.Response{
			Code:    http.StatusForbidden,
			Message: "初始化管理员权限组不可修改",
		})
		return
	}

	if adminID != initialAdminUserID {
		c.JSON(http.StatusForbidden, dto.Response{
			Code:    http.StatusForbidden,
			Message: "仅初始化管理员可修改用户权限组",
		})
		return
	}

	var targetUser models.User
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "开启事务失败：" + tx.Error.Error(),
		})
		return
	}

	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&targetUser, req.UserID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{
				Code:    http.StatusNotFound,
				Message: "目标用户不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "数据库查询失败：" + err.Error(),
		})
		return
	}

	if err := tx.Model(&targetUser).Update("group", req.NewGroup).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "修改权限组失败：" + err.Error(),
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "提交修改请求失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "修改用户权限组成功",
	})
}

// @Summary		管理员按类型查询基础资源
// @Description	通过query参数type查询skills/achievements/items/cards；支持分页与可选id精确查询
// @Description	type可选值：achievements、skills、items、cards
// @Description	type=achievements 时，data.list 为 []models.Achievement
// @Description	type=skills 时，data.list 为 []models.Skill
// @Description	type=items 时，data.list 为 []models.Item
// @Description	type=cards 时，data.list 为 []models.Card
// @Tags			admin-get
// @Produce		json
// @Param			type		query		string									true	"资源类型(achievements/skills/items/cards)"
// @Param			id			query		int										false	"资源ID，传入后优先精确查询"
// @Param			name		query		string									false	"名称模糊搜索"
// @Param			skill_group	query		string									false	"技能组模糊搜索(type=skills有效)"
// @Param			page		query		int										false	"页码，默认1"
// @Param			page_size	query		int										false	"每页多少，默认20，最大100"
// @Success		200			{object}	dto.Response{data=dto.PaginatedData}	"type=achievements 查询成功"
// @Success		200			{object}	dto.Response{data=dto.PaginatedData}	"type=skills 查询成功"
// @Success		200			{object}	dto.Response{data=dto.PaginatedData}	"type=items 查询成功"
// @Success		200			{object}	dto.Response{data=dto.PaginatedData}	"type=cards 查询成功"
// @Failure		400			{object}	dto.Response							"请求参数错误"
// @Failure		401			{object}	dto.Response							"登录状态异常"
// @Failure		500			{object}	dto.Response							"数据库查询失败"
// @Router			/api/admin/get/resources [get]
func GetResourcesByTypeForAdmin(c *gin.Context) {
	relationType, err := parseResourceType(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	handler, exists := adminResourceQueryHandlers[relationType]
	if !exists {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: ErrUnsupportedRelationType.Error()})
		return
	}

	pagination := middleware.GetPagination(c)
	list, total, err := handler(c, pagination)
	if err != nil {
		if errors.Is(err, ErrResourceIDInvalid) {
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "数据库查询失败：" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "查询成功",
		Data: dto.PaginatedData{
			List:     list,
			Total:    total,
			Page:     pagination.Page,
			PageSize: pagination.PageSize,
		},
	})
}

// @Summary		管理员按类型创建基础资源
// @Description	通过query参数type创建skills/achievements/items/cards中的一种资源
// @Description	type=achievements 请求体：dto.AdminCreateAchievementRequest
// @Description	type=skills 请求体：dto.AdminCreateSkillRequest
// @Description	type=items 请求体：dto.AdminCreateItemRequest
// @Description	type=cards 请求体：dto.AdminCreateCardRequest
// @Description	type=achievements 时，data 为 models.Achievement
// @Description	type=skills 时，data 为 models.Skill
// @Description	type=items 时，data 为 models.Item
// @Description	type=cards 时，data 为 models.Card
// @Tags			admin-operation
// @Accept			json
// @Produce		json
// @Param			type	query		string									true	"资源类型(achievements/skills/items/cards)"
// @Param			request	body		dto.AdminCreateSkillRequest				true	"创建请求体(示例以skills为准)"
// @Success		200		{object}	dto.Response{data=models.Achievement}	"type=achievements 创建成功"
// @Success		200		{object}	dto.Response{data=models.Skill}			"type=skills 创建成功"
// @Success		200		{object}	dto.Response{data=models.Item}			"type=items 创建成功"
// @Success		200		{object}	dto.Response{data=models.Card}			"type=cards 创建成功"
// @Failure		400		{object}	dto.Response							"请求参数错误"
// @Failure		401		{object}	dto.Response							"登录状态异常"
// @Failure		409		{object}	dto.Response							"名称冲突"
// @Failure		500		{object}	dto.Response							"数据库错误"
// @Router			/api/admin/operation/resources [post]
func CreateResourceByTypeForAdmin(c *gin.Context) {
	relationType, err := parseResourceType(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	handler, exists := adminResourceCreateHandlers[relationType]
	if !exists {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: ErrUnsupportedRelationType.Error()})
		return
	}

	resource, err := handler(c)
	if err != nil {
		if errors.Is(err, ErrInvalidRequestBody) {
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
			return
		}
		if errors.Is(err, ErrResourceNameExists) {
			c.JSON(http.StatusConflict, dto.Response{Code: http.StatusConflict, Message: err.Error()})
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "目标资源不存在"})
			return
		}
		if errors.Is(err, ErrNoUpdateFields) {
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "创建失败：" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "创建成功", Data: resource})
}

// @Summary		管理员按类型更新基础资源
// @Description	通过query参数type更新skills/achievements/items/cards中的一种资源
// @Description	type=achievements 请求体：dto.AdminUpdateAchievementRequest
// @Description	type=skills 请求体：dto.AdminUpdateSkillRequest
// @Description	type=items 请求体：dto.AdminUpdateItemRequest
// @Description	type=cards 请求体：dto.AdminUpdateCardRequest
// @Description	type=achievements 时，data 为 models.Achievement
// @Description	type=skills 时，data 为 models.Skill
// @Description	type=items 时，data 为 models.Item
// @Description	type=cards 时，data 为 models.Card
// @Tags			admin-update
// @Accept			json
// @Produce		json
// @Param			type	query		string									true	"资源类型(achievements/skills/items/cards)"
// @Param			request	body		dto.AdminUpdateSkillRequest				true	"更新请求体(示例以skills为准)"
// @Success		200		{object}	dto.Response{data=models.Achievement}	"type=achievements 更新成功"
// @Success		200		{object}	dto.Response{data=models.Skill}			"type=skills 更新成功"
// @Success		200		{object}	dto.Response{data=models.Item}			"type=items 更新成功"
// @Success		200		{object}	dto.Response{data=models.Card}			"type=cards 更新成功"
// @Failure		400		{object}	dto.Response							"请求参数错误"
// @Failure		401		{object}	dto.Response							"登录状态异常"
// @Failure		404		{object}	dto.Response							"目标资源不存在"
// @Failure		409		{object}	dto.Response							"名称冲突"
// @Failure		500		{object}	dto.Response							"数据库错误"
// @Router			/api/admin/update/resources [put]
func UpdateResourceByTypeForAdmin(c *gin.Context) {
	relationType, err := parseResourceType(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	handler, exists := adminResourceUpdateHandlers[relationType]
	if !exists {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: ErrUnsupportedRelationType.Error()})
		return
	}

	resource, err := handler(c)
	if err != nil {
		if errors.Is(err, ErrInvalidRequestBody) {
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "目标资源不存在"})
			return
		}
		if errors.Is(err, ErrResourceNameExists) {
			c.JSON(http.StatusConflict, dto.Response{Code: http.StatusConflict, Message: err.Error()})
			return
		}
		if errors.Is(err, ErrNoUpdateFields) {
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "更新失败：" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "更新成功", Data: resource})
}

// @Summary		管理员按类型删除基础资源
// @Description	通过query参数type删除skills/achievements/items/cards中的一种资源
// @Tags			admin-operation
// @Accept			json
// @Produce		json
// @Param			type	query		string									true	"资源类型(achievements/skills/items/cards)"
// @Param			request	body		dto.AdminDeleteResourceByTypeRequest	true	"删除请求体"
// @Success		200		{object}	dto.Response							"删除成功"
// @Failure		400		{object}	dto.Response							"请求参数错误"
// @Failure		401		{object}	dto.Response							"登录状态异常"
// @Failure		404		{object}	dto.Response							"目标资源不存在"
// @Failure		500		{object}	dto.Response							"数据库错误"
// @Router			/api/admin/operation/resources [delete]
func DeleteResourceByTypeForAdmin(c *gin.Context) {
	relationType, err := parseResourceType(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	var req dto.AdminDeleteResourceByTypeRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "请求参数错误:" + err.Error()})
		return
	}

	handler, exists := adminResourceDeleteHandlers[relationType]
	if !exists {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: ErrUnsupportedRelationType.Error()})
		return
	}

	if err = handler(req.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "目标资源不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "删除失败：" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "删除成功"})
}

// @Summary	获取用户自身基础信息
// @Tags		profile-get
// @Produce	json
// @Success	200	{object}	dto.Response{data=dto.CommonUserData}	"查询成功"
// @Failure	401	{object}	dto.Response							"登录状态异常"
// @Failure	500	{object}	dto.Response							"数据库查询失败"
// @Router		/api/profile/get/self [get]
func GetSelfProfile(c *gin.Context) {
	var err error
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized, //401
			Message: "解析后token中缺少用户信息",
		})
		return
	}
	var user models.User

	err = config.DB.First(&user, userId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{
				Code:    http.StatusNotFound, //404
				Message: "用户不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, dto.Response{
				Code:    http.StatusInternalServerError, //500
				Message: "数据库查询失败：" + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "查询成功",
		Data: dto.CommonUserData{
			UserID:        user.ID,
			Username:      user.Username,
			Group:         user.Group,
			HeadImagePath: user.HeadImagePath,
			StrengthCoin:  user.StrengthCoin,
			SelectCoin:    user.SelectCoin,
		},
	})
}

// @Summary		按类型获取用户关联数据
// @Description	通过query参数type选择查询achievements/skills/items/cards中的一种
// @Description	type=achievements 时，data.list 为 []dto.UserAchievementRelationData
// @Description	type=skills 时，data.list 为 []dto.UserSkillRelationData
// @Description	type=items 时，data.list 为 []dto.UserItemRelationData
// @Description	type=cards 时，data.list 为 []dto.UserCardRelationData
// @Tags			profile-get
// @Produce		json
// @Param			type		query		string													true	"关联类型(achievements/skills/items/cards)"
// @Param			page		query		int														false	"页码，默认1"
// @Param			page_size	query		int														false	"每页多少，默认20，最大100"
// @Success		200			{object}	dto.Response{data=dto.UserAchievementRelationPageData}	"type=achievements 查询成功"
// @Success		200			{object}	dto.Response{data=dto.UserSkillRelationPageData}		"type=skills 查询成功"
// @Success		200			{object}	dto.Response{data=dto.UserItemRelationPageData}			"type=items 查询成功"
// @Success		200			{object}	dto.Response{data=dto.UserCardRelationPageData}			"type=cards 查询成功"
// @Failure		400			{object}	dto.Response											"请求参数错误"
// @Failure		401			{object}	dto.Response											"登录状态异常"
// @Failure		500			{object}	dto.Response											"数据库查询失败"
// @Router			/api/profile/get/relations [get]
func GetSelfRelationsByType(c *gin.Context) {
	relationTypeStr := c.Query("type")
	if relationTypeStr == "" {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "缺少type参数",
		})
		return
	}

	relationType, err := ParseUserRelationType(relationTypeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	list, total, err := QueryUserRelationByType(c, relationType)
	if err != nil {
		if errors.Is(err, ErrUserIDMissing) || errors.Is(err, ErrUserIDTypeInvalid) {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
			})
			return
		}
		if errors.Is(err, ErrUnsupportedRelationType) {
			c.JSON(http.StatusBadRequest, dto.Response{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "数据库查询失败：" + err.Error(),
		})
		return
	}

	pagination := middleware.GetPagination(c)
	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "查询成功",
		Data: dto.PaginatedData{
			List:     list,
			Total:    total,
			Page:     pagination.Page,
			PageSize: pagination.PageSize,
		},
	})
}
