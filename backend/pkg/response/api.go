package response

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	"gkube/pkg/logger"
)

// 响应契约:
//   - 成功: code=200, HTTP 200
//   - 业务失败: code=1001-1004(参数/认证/权限/资源) + 4xx HTTP
//   - K8s 错误: code=1005(客户端获取失败)/1006(API调用失败) + 502 HTTP
//   - 兜底错误: code=1007 + 500 HTTP（仅未包装为 AppError 的未知错误）
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
	safe := strings.ReplaceAll(strings.ReplaceAll(filename, "\n", ""), "\r", "")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, safe))
	c.Data(http.StatusOK, "application/octet-stream", res)
}

// FailWithError 根据 error 类型自动选择响应格式。
//   - *apperr.AppError → 使用其 HTTPStatus 和 Code，msg 脱敏，原始错误写日志
//   - 其他 error → HTTP 500 + code=1007，原始错误写日志
func FailWithError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		if appErr.Err != nil {
			logger.Error(fmt.Sprintf("%s: %s", appErr.Message, appErr.Err.Error()))
		} else {
			logger.Error(appErr.Message)
		}
		c.JSON(appErr.HTTPStatus, gin.H{
			"msg":  appErr.Message,
			"code": int(appErr.Code),
			"data": nil,
		})
		c.Abort()
		return
	}
	// 未知错误，兜底 500
	logger.Error(fmt.Sprintf("未知错误: %s", err.Error()))
	c.JSON(http.StatusInternalServerError, gin.H{
		"msg":  "服务器内部错误",
		"code": int(apperr.ErrCodeInternal),
		"data": nil,
	})
	c.Abort()
}
