package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

// Ok 成功响应
func Ok(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Data: nil,
		Msg:  "success",
	})
}

// OkWithData 成功响应带数据
func OkWithData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Data: data,
		Msg:  "success",
	})
}

// OkWithDetailed 成功响应带详细信息
func OkWithDetailed(c *gin.Context, data interface{}, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Data: data,
		Msg:  msg,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: 1,
		Data: nil,
		Msg:  msg,
	})
}

// FailWithCode 失败响应带错误码
func FailWithCode(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Data: nil,
		Msg:  msg,
	})
}
