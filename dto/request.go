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
