package dto

import "mime/multipart"

// @summary		用户注册请求
// @description	注册信息
type RegisterRequest struct {
	UserName string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,max=25"`
}

// @summary		用户登录请求
// @description	登录信息
type LoginRequest struct {
	UserName string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// @summary		用户修改密码
// @description	修改密码
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=25"`
}

// @summary		用户修改用户名
// @description	修改用户名
type UpdateUsernameRequest struct {
	NewUsername string `json:"new_username" binding:"required,min=3,max=20"`
}

// @summary		用户修改头像
// @description	用户修改头像
type UpdateHeadImageRequest struct {
	NewHeadImage *multipart.FileHeader `form:"new_head_image" binding:"required"`
}

// @summary		管理员删除用户请求
// @description	按用户ID删除用户
type AdminDeleteUserRequest struct {
	UserID uint `json:"user_id" binding:"required,gt=0"`
}

// @summary		管理员修改用户权限组请求
// @description	按用户ID修改权限组，仅支持user/admin
type AdminUpdateUserGroupRequest struct {
	UserID   uint   `json:"user_id" binding:"required,gt=0"`
	NewGroup string `json:"new_group" binding:"required,oneof=user admin"`
}

type AdminCreateAchievementRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Description string `json:"description"`
}

type AdminCreateSkillRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Description string `json:"description"`
	SkillGroup  string `json:"skill_group"`
	PrqSkillID  uint   `json:"prq_skill_id"`
}

type AdminCreateItemRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Description string `json:"description"`
}

type AdminCreateCardRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Description string `json:"description"`
}

// @summary		通用资源创建请求
// @description	用于achievements/items/cards，skills可额外携带技能字段
type CommonResourceCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SkillGroup  string `json:"skill_group,omitempty"`
	PrqSkillID  uint   `json:"prq_skill_id,omitempty"`
}

type AdminUpdateAchievementRequest struct {
	ID          uint    `json:"id" binding:"required,gt=0"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type AdminUpdateSkillRequest struct {
	ID          uint    `json:"id" binding:"required,gt=0"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	SkillGroup  *string `json:"skill_group"`
	PrqSkillID  *uint   `json:"prq_skill_id"`
}

type AdminUpdateItemRequest struct {
	ID          uint    `json:"id" binding:"required,gt=0"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type AdminUpdateCardRequest struct {
	ID          uint    `json:"id" binding:"required,gt=0"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// @summary		通用资源更新请求
// @description	用于achievements/items/cards，skills可额外携带技能字段
type CommonResourceUpdateRequest struct {
	ID          uint    `json:"id"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	SkillGroup  *string `json:"skill_group,omitempty"`
	PrqSkillID  *uint   `json:"prq_skill_id,omitempty"`
}

// @summary		管理员按类型删除基础资源请求
// @description	用于skills/achievements/items/cards的删除（按ID）
type AdminDeleteResourceByTypeRequest struct {
	ID uint `json:"id" binding:"required,gt=0"`
}
