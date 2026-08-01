package response

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gkube/pkg/logger"
)

// 响应契约:
//   - 成功: code=200, HTTP 200
//   - 业务失败: code=0(非 200) + 4xx HTTP
//   - 服务端/DB/k8s 错误: code=0(非 200) + 5xx HTTP
//   - 对外 msg 脱敏,不回显底层错误原文;原始错误写入 logger。

// Success 成功响应,code=200。
func Success(c *gin.Context, msg string, data any) {
	c.JSON(http.StatusOK, gin.H{
		"msg":  msg,
		"code": 200,
		"data": data,
	})
}

// Fail 业务失败响应,默认 HTTP 400 + code=0。msg 应为脱敏后的用户可读文案。
func Fail(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"msg":  msg,
		"code": 0,
		"data": nil,
	})
	c.Abort()
}

// FailWithStatus 返回指定 HTTP 状态码的错误响应(code=0)。
func FailWithStatus(c *gin.Context, statusCode int, msg string) {
	c.JSON(statusCode, gin.H{
		"msg":  msg,
		"code": 0,
		"data": nil,
	})
	c.Abort()
}

// FailWithLog 记录原始错误到日志并返回脱敏的错误响应。
// statusCode: 业务错误用 4xx,服务端/DB/k8s 错误用 5xx。msg 为脱敏文案。
func FailWithLog(c *gin.Context, statusCode int, msg string, err error) {
	if err != nil {
		logger.Error(fmt.Sprintf("%s: %s", msg, err.Error()))
	}
	FailWithStatus(c, statusCode, msg)
}

// FailServer 服务端错误便捷方法,HTTP 500 + 脱敏文案 + 记录原始错误。
func FailServer(c *gin.Context, msg string, err error) {
	FailWithLog(c, http.StatusInternalServerError, msg, err)
}

// File 文件响应
func File(c *gin.Context, filename string, res []byte) {
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/octet-stream", res)
}
