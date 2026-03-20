package handler

import (
	config "MuXi/2026-MuxiShooter-Backend/config"
	"MuXi/2026-MuxiShooter-Backend/dto"
	"MuXi/2026-MuxiShooter-Backend/middleware"
	"MuXi/2026-MuxiShooter-Backend/service"
	"MuXi/2026-MuxiShooter-Backend/utils"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProfileHandler struct {
	profileService *service.ProfileService
}

func NewProfileHandler(profileService *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

// Logout godoc
// @Summary      用户登出
// @Description  当前登录用户登出（使现有token失效）
// @Tags         profile
// @Produce      json
// @Success      200  {object}  dto.Response  "登出成功"
// @Failure      401  {object}  dto.Response  "登录状态异常或用户不存在"
// @Failure      500  {object}  dto.Response  "服务器错误"
// @Security     BearerAuth
// @Router       /profile/operation/logout [get]
func (h *ProfileHandler) Logout(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: service.ErrMissingUserContext.Error(),
		})
		return
	}

	err := h.profileService.Logout(userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized,
				Message: "用户不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("服务器错误: %v, 请重试", err),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "登出成功",
	})
}

// UpdatePassword godoc
// @Summary      用户修改密码
// @Description  通过旧密码校验后更新密码，成功后当前登录状态失效
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        body  body      dto.UpdatePasswordRequest  true  "修改密码请求体"
// @Success      200   {object}  dto.Response               "修改密码成功"
// @Failure      400   {object}  dto.Response               "请求参数错误"
// @Failure      401   {object}  dto.Response               "登录状态异常"
// @Failure      403   {object}  dto.Response               "旧密码错误或修改过于频繁"
// @Failure      404   {object}  dto.Response               "用户不存在"
// @Failure      500   {object}  dto.Response               "服务器错误"
// @Security     BearerAuth
// @Router       /profile/update/password [put]
func (h *ProfileHandler) UpdatePassword(c *gin.Context) {
	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误:" + err.Error(),
		})
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: service.ErrMissingUserContext.Error(),
		})
		return
	}

	err := h.profileService.UpdatePassword(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSamePassword):
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		case errors.Is(err, service.ErrUserNotFound), errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "用户不存在"})
		case errors.Is(err, service.ErrPasswordTooFrequent):
			c.JSON(http.StatusForbidden, dto.Response{Code: http.StatusForbidden, Message: err.Error()})
		case errors.Is(err, service.ErrInvalidOldPassword):
			c.JSON(http.StatusForbidden, dto.Response{Code: http.StatusForbidden, Message: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "服务器错误:" + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "修改密码成功，已退出登录，请重新登陆",
	})
}

// UpdateUsername godoc
// @Summary      用户修改用户名
// @Description  修改当前用户用户名
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        body  body      dto.UpdateUsernameRequest  true  "修改用户名请求体"
// @Success      200   {object}  dto.Response               "修改成功"
// @Failure      400   {object}  dto.Response               "请求参数错误"
// @Failure      401   {object}  dto.Response               "登录状态异常"
// @Failure      403   {object}  dto.Response               "修改过于频繁"
// @Failure      404   {object}  dto.Response               "用户不存在"
// @Failure      500   {object}  dto.Response               "服务器错误"
// @Security     BearerAuth
// @Router       /profile/update/username [put]
func (h *ProfileHandler) UpdateUsername(c *gin.Context) {
	var req dto.UpdateUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误:" + err.Error(),
		})
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: service.ErrMissingUserContext.Error(),
		})
		return
	}

	err := h.profileService.UpdateUsername(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound), errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "用户不存在"})
		case errors.Is(err, service.ErrUsernameTooFrequent):
			c.JSON(http.StatusForbidden, dto.Response{Code: http.StatusForbidden, Message: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "服务器错误:" + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "修改用户名成功",
	})
}

// UpdateHeadImage godoc
// @Summary      用户修改头像
// @Description  上传并更新当前用户头像
// @Tags         profile
// @Accept       multipart/form-data
// @Produce      json
// @Param        new_head_image  formData  file  true  "头像文件"
// @Success      200  {object}  dto.Response  "修改成功"
// @Failure      400  {object}  dto.Response  "请求参数错误"
// @Failure      401  {object}  dto.Response  "登录状态异常"
// @Failure      403  {object}  dto.Response  "修改过于频繁"
// @Failure      404  {object}  dto.Response  "用户不存在"
// @Failure      500  {object}  dto.Response  "服务器错误"
// @Security     BearerAuth
// @Router       /profile/update/headimage [put]
func (h *ProfileHandler) UpdateHeadImage(c *gin.Context) {
	var req dto.UpdateHeadImageRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误:" + err.Error(),
		})
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: service.ErrMissingUserContext.Error(),
		})
		return
	}

	if req.NewHeadImage == nil || req.NewHeadImage.Size == 0 {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "头像为空",
		})
		return
	}

	log.Printf("用户(id:%d)上传头像,Size:%d", userID, req.NewHeadImage.Size)
	savePath, err := utils.SaveImages(c, req.NewHeadImage, config.PrefixHeadImg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "图片保存失败：" + err.Error(),
		})
		return
	}

	oldHeadImagePath, err := h.profileService.UpdateHeadImage(userID, savePath)
	if err != nil {
		_ = utils.RemoveFile(savePath)
		switch {
		case errors.Is(err, service.ErrUserNotFound), errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "用户不存在"})
		case errors.Is(err, service.ErrHeadImageTooFrequent):
			c.JSON(http.StatusForbidden, dto.Response{Code: http.StatusForbidden, Message: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "修改头像失败：" + err.Error()})
		}
		return
	}

	if oldHeadImagePath != "" && oldHeadImagePath != savePath && oldHeadImagePath != config.DefaultHeadImagePath {
		if removeErr := utils.RemoveFile(oldHeadImagePath); removeErr != nil {
			log.Printf("删除旧头像失败(user_id:%d,path:%s): %v", userID, oldHeadImagePath, removeErr)
		}
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "修改头像成功",
	})
}

// UpdateCoinByType godoc
// @Summary      用户按类型修改金币
// @Description  通过query参数type(strength/select)修改对应金币
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        type  query     string                   true  "金币类型(strength/select)"
// @Param        body  body      dto.UpdateCoinByTypeRequest true  "修改金币请求体"
// @Success      200   {object}  dto.Response             "修改成功"
// @Failure      400   {object}  dto.Response             "请求参数错误"
// @Failure      401   {object}  dto.Response             "登录状态异常"
// @Failure      404   {object}  dto.Response             "用户不存在"
// @Failure      500   {object}  dto.Response             "数据库错误"
// @Security     BearerAuth
// @Router       /profile/update/coin [put]
func (h *ProfileHandler) UpdateCoinByType(c *gin.Context) {
	var req dto.UpdateCoinByTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误:" + err.Error(),
		})
		return
	}
	if req.Coin == nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "coin不能为空"})
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: service.ErrMissingUserContext.Error(),
		})
		return
	}

	coinType := c.Query("type")
	userData, err := h.profileService.UpdateCoinByType(userID, coinType, *req.Coin)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMissingCoinType), errors.Is(err, service.ErrUnsupportedCoinType):
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		case errors.Is(err, service.ErrUserNotFound), errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "用户不存在"})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "数据库查询失败：" + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "修改金币成功",
		Data:    userData,
	})
}

// CreateSelfRelationByType godoc
// @Summary      用户按类型创建自身资源关联
// @Description  通过query参数type(achievements/skills/items/cards)和body中的resource_id创建本人关联记录
// @Tags         profile-relation
// @Accept       json
// @Produce      json
// @Param        type  query     string                       true  "关联类型(achievements/skills/items/cards)"
// @Param        body  body      dto.UserRelationCreateRequest true  "创建关联请求体"
// @Success      200   {object}  dto.Response                 "创建成功"
// @Failure      400   {object}  dto.Response                 "请求参数错误"
// @Failure      401   {object}  dto.Response                 "登录状态异常"
// @Failure      404   {object}  dto.Response                 "目标资源不存在"
// @Failure      409   {object}  dto.Response                 "关联已存在"
// @Failure      500   {object}  dto.Response                 "创建失败"
// @Security     BearerAuth
// @Router       /profile/operation/relations [post]
func (h *ProfileHandler) CreateSelfRelationByType(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{Code: http.StatusUnauthorized, Message: service.ErrMissingUserContext.Error()})
		return
	}

	relationTypeStr := c.Query("type")
	if relationTypeStr == "" {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "缺少type参数"})
		return
	}
	relationType, err := service.ParseUserRelationType(relationTypeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	var req dto.UserRelationCreateRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "请求参数错误:" + err.Error()})
		return
	}

	log.Printf("创建关联请求: user_id=%d type=%s resource_id=%d", userID, relationType, req.ResourceID)

	data, err := h.profileService.CreateSelfRelationByType(userID, relationType, req.ResourceID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "目标资源不存在"})
		case errors.Is(err, service.ErrResourceNameExists):
			c.JSON(http.StatusConflict, dto.Response{Code: http.StatusConflict, Message: "关联已存在"})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "创建失败：" + err.Error()})
		}
		return
	}

	if data.Resource.ResourceID == 0 || data.Resource.ResourceID != req.ResourceID {
		log.Printf("创建关联结果异常: user_id=%d type=%s req_resource_id=%d resp_resource_id=%d", userID, relationType, req.ResourceID, data.Resource.ResourceID)
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: service.ErrRelationCreateInconsistent.Error()})
		return
	}

	log.Printf("创建关联成功: user_id=%d type=%s resource_id=%d", userID, relationType, req.ResourceID)

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "创建成功", Data: data})
}

// UpdateSelfRelationByType godoc
// @Summary      用户按类型更新自身资源关联
// @Description  通过query参数type(achievements/skills/items/cards)和body更新本人关联记录，skills支持skill_grade
// @Tags         profile-relation
// @Accept       json
// @Produce      json
// @Param        type  query     string                       true  "关联类型(achievements/skills/items/cards)"
// @Param        body  body      dto.UserRelationUpdateRequest true  "更新关联请求体"
// @Success      200   {object}  dto.Response                 "更新成功"
// @Failure      400   {object}  dto.Response                 "请求参数错误"
// @Failure      401   {object}  dto.Response                 "登录状态异常"
// @Failure      404   {object}  dto.Response                 "关联记录不存在"
// @Failure      500   {object}  dto.Response                 "更新失败"
// @Security     BearerAuth
// @Router       /profile/update/relations [put]
func (h *ProfileHandler) UpdateSelfRelationByType(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{Code: http.StatusUnauthorized, Message: service.ErrMissingUserContext.Error()})
		return
	}

	relationTypeStr := c.Query("type")
	if relationTypeStr == "" {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "缺少type参数"})
		return
	}
	relationType, err := service.ParseUserRelationType(relationTypeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	var req dto.UserRelationUpdateRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "请求参数错误:" + err.Error()})
		return
	}

	data, err := h.profileService.UpdateSelfRelationByType(userID, relationType, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNoUpdateFields), errors.Is(err, service.ErrSkillGradeOnlyForSkills), errors.Is(err, service.ErrUnsupportedRelationType):
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "关联记录不存在"})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "更新失败：" + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "更新成功", Data: data})
}

// DeleteSelfRelationByType godoc
// @Summary      用户按类型删除自身资源关联
// @Description  通过query参数type(achievements/skills/items/cards)和body中的resource_id删除本人关联记录
// @Tags         profile-relation
// @Accept       json
// @Produce      json
// @Param        type  query     string                       true  "关联类型(achievements/skills/items/cards)"
// @Param        body  body      dto.UserRelationDeleteRequest true  "删除关联请求体"
// @Success      200   {object}  dto.Response                 "删除成功"
// @Failure      400   {object}  dto.Response                 "请求参数错误"
// @Failure      401   {object}  dto.Response                 "登录状态异常"
// @Failure      404   {object}  dto.Response                 "关联记录不存在"
// @Failure      500   {object}  dto.Response                 "删除失败"
// @Security     BearerAuth
// @Router       /profile/operation/relations [delete]
func (h *ProfileHandler) DeleteSelfRelationByType(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{Code: http.StatusUnauthorized, Message: service.ErrMissingUserContext.Error()})
		return
	}

	relationTypeStr := c.Query("type")
	if relationTypeStr == "" {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "缺少type参数"})
		return
	}
	relationType, err := service.ParseUserRelationType(relationTypeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	var req dto.UserRelationDeleteRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "请求参数错误:" + err.Error()})
		return
	}

	if err = h.profileService.DeleteSelfRelationByType(userID, relationType, req.ResourceID); err != nil {
		switch {
		case errors.Is(err, service.ErrUnsupportedRelationType):
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "关联记录不存在"})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "删除失败：" + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "删除成功"})
}

// GetSelfProfile godoc
// @Summary      获取当前用户信息
// @Description  查询当前登录用户的基础信息
// @Tags         profile
// @Produce      json
// @Success      200  {object}  dto.Response  "查询成功"
// @Failure      401  {object}  dto.Response  "登录状态异常"
// @Failure      404  {object}  dto.Response  "用户不存在"
// @Failure      500  {object}  dto.Response  "数据库错误"
// @Security     BearerAuth
// @Router       /profile/get/self [get]
func (h *ProfileHandler) GetSelfProfile(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{Code: http.StatusUnauthorized, Message: service.ErrMissingUserContext.Error()})
		return
	}

	data, err := h.profileService.GetSelfProfile(userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound), errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "用户不存在"})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "数据库查询失败：" + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "查询成功", Data: data})
}

// GetSelfRelationsByType godoc
// @Summary      用户按类型查询自身资源关联
// @Description  通过query参数type查询本人在achievements/skills/items/cards中的关联数据，支持分页
// @Tags         profile-relation
// @Produce      json
// @Param        type       query     string  true   "关联类型(achievements/skills/items/cards)"
// @Param        page       query     int     false  "页码，默认1"
// @Param        page_size  query     int     false  "每页数量，默认20，最大100"
// @Success      200        {object}  dto.Response  "查询成功"
// @Failure      400        {object}  dto.Response  "请求参数错误"
// @Failure      401        {object}  dto.Response  "登录状态异常"
// @Failure      500        {object}  dto.Response  "查询失败"
// @Security     BearerAuth
// @Router       /profile/get/relations [get]
func (h *ProfileHandler) GetSelfRelationsByType(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{Code: http.StatusUnauthorized, Message: service.ErrMissingUserContext.Error()})
		return
	}

	pagination := middleware.GetPagination(c)
	list, total, err := h.profileService.GetSelfRelationsByType(userID, c.Query("type"), pagination)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMissingRelationType), errors.Is(err, service.ErrUnsupportedRelationType):
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "数据库查询失败：" + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "查询成功",
		Data: dto.CommonUserRelationPageData{
			List:     list,
			Total:    total,
			Page:     pagination.Page,
			PageSize: pagination.PageSize,
		},
	})
}

func getUserIDFromContext(c *gin.Context) (uint, bool) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	userID, ok := userIDValue.(uint)
	if !ok {
		return 0, false
	}
	return userID, true
}
