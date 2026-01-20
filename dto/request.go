package dto

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
