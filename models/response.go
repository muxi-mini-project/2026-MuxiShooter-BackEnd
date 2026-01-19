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
