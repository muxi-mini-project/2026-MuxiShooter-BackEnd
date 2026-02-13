package dto

// @description	通用响应结构体
type Response struct {
	//http状态码
	Code int `json:"code"`
	//返回的消息
	Message string `json:"message"`
	//根据具体数据来定类型
	Data interface{} `json:"data"`
}

type CommonUserData struct {
	//用户ID
	UserID uint `json:"user_id"`
	//用户名
	Username string `json:"username"`
	//权限组
	Group string `json:"group"`
	//头像路径
	HeadImagePath string `json:"head_image_path"`
	//强化货币
	StrengthCoin uint `json:"strength_coin"`
	//抽卡货币
	SelectCoin uint `json:"select_coin"`
}

type AuthData struct {
	//用户
	User CommonUserData `json:"user"`
	//JWT Token
	Token string `json:"token"`
	//有效期，24h
	ExpiresAt int64 `json:"expires_at"`
}

type PaginatedData struct {
	//查询所得列表
	List interface{} `json:"list"`
	//返回数据总数
	Total int64 `json:"total"`
	//页码
	Page int `json:"page"`
	//每页多少
	PageSize int `json:"page_size"`
}
