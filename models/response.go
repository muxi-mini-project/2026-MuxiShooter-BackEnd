package models

// @description	通用响应结构体
type Response struct {
	//http状态码
	Code int `json:"code"`
	//返回的消息
	Message string `json:"message"`
	//根据具体数据来定类型
	Data interface{} `json:"data"`
}

// @description	注册返回Data结构体
type RegisterData struct {
	Name string `json:"username"`
	//用户ID minimum(1)
	UserID uint `json:"user_id"`
}

// @description	登录返回Data结构体
type LoginData struct {
	Name string `json:"username"`
	//user or admin
	UserGroup uint `json:"user_group"`
}
