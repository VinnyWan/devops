package common

import "github.com/gin-gonic/gin"

// Response 通用响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// PageResult 分页响应结构
type PageResult struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code: 200,
		Msg:  "操作成功",
		Data: data,
	})
}

// SuccessWithMsg 成功响应(自定义消息)
func SuccessWithMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(200, Response{
		Code: 200,
		Msg:  msg,
		Data: data,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, msg string) {
	c.JSON(200, Response{
		Code: 500,
		Msg:  msg,
	})
}

// FailWithCode 失败响应(自定义状态码)
func FailWithCode(c *gin.Context, code int, msg string) {
	c.JSON(200, Response{
		Code: code,
		Msg:  msg,
	})
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(401, Response{
		Code: 401,
		Msg:  msg,
	})
	c.Abort()
}

// Forbidden 禁止访问响应
func Forbidden(c *gin.Context, msg string) {
	c.JSON(403, Response{
		Code: 403,
		Msg:  msg,
	})
	c.Abort()
}

// PageSuccess 分页成功响应
func PageSuccess(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(200, Response{
		Code: 200,
		Msg:  "操作成功",
		Data: PageResult{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}
