package models

import "mime/multipart"

// @summary		用户注册请求
// @description	注册信息
type RegisterRequest struct {
	Name string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,max=25"`
}

// @summary		用户登录请求
// @description	登录信息
type LoginRequest struct {
	Name string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// @summary		创建图书请求
// @description	创建图书所需参数,由于包含封面图片文件，所以用form
type CreateBookRequest struct {
	Title        string                `form:"title" binding:"required"`
	Author       string                `form:"author" binding:"required"`
	Summary      string                `form:"summary" binding:"omitempty"`
	Cover        *multipart.FileHeader `form:"cover" binding:"omitempty"`
	InitialStock int                   `form:"initial_stock" binding:"gte=0"`
}

// @summary		图书查询请求(ID)
// @description	包含图书ID的请求
type FindBookRequest struct {
	//图书ID minimum(1)
	ID uint `json:"book_id" binding:"required"`
}

// @summary		更新图书请求
// @description	更新图书请求所需参数
type UpdateBookRequest struct {
	ID         uint                  `form:"book_id"`
	Title      string                `form:"title" binding:"omitempty"`
	Author     string                `form:"author" binding:"omitempty"`
	Summary    string                `form:"summary" binding:"omitempty"`
	Cover      *multipart.FileHeader `form:"cover" binding:"omitempty"`
	Stock      int                   `form:"stock" binding:"gte=0"`
	TotalStock int                   `form:"total_stock" binding:"gte=0"`
}
