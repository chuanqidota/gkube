package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// 标准库 errors 的常用函数透传，方便统一从本包导入
var (
	As = errors.As
	Is = errors.Is
	New = errors.New
)

// ErrorCode 业务错误码，用于前端程序化区分错误类型
type ErrorCode int

const (
	ErrCodeValidation ErrorCode = 1001 // 参数校验失败
	ErrCodeNotFound   ErrorCode = 1002 // 资源不存在
	ErrCodeForbidden  ErrorCode = 1003 // 无权限
	ErrCodeConflict   ErrorCode = 1004 // 资源冲突（409）
	ErrCodeK8sClient  ErrorCode = 1005 // K8s 客户端创建失败
	ErrCodeK8sAPI     ErrorCode = 1006 // K8s API 调用失败
	ErrCodeInternal   ErrorCode = 1007 // 内部错误
)

// AppError 应用级错误，携带错误码、HTTP 状态码、脱敏消息和原始错误
type AppError struct {
	Code       ErrorCode // 业务错误码，写入 JSON body 的 code 字段
	HTTPStatus int       // HTTP 状态码
	Message    string    // 对外脱敏消息，写入 JSON body 的 msg 字段
	Err        error     // 内部原始错误，仅写入日志，不返回给客户端
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// ---------- 构造函数 ----------

// Validation 参数校验失败，HTTP 400
func Validation(msg string, err error) *AppError {
	return &AppError{Code: ErrCodeValidation, HTTPStatus: http.StatusBadRequest, Message: msg, Err: err}
}

// BadRequest 请求参数错误，HTTP 400（Validation 的语义别名）
func BadRequest(msg string, err error) *AppError {
	return &AppError{Code: ErrCodeValidation, HTTPStatus: http.StatusBadRequest, Message: msg, Err: err}
}

// NotFound 资源不存在，HTTP 404
func NotFound(msg string, err error) *AppError {
	return &AppError{Code: ErrCodeNotFound, HTTPStatus: http.StatusNotFound, Message: msg, Err: err}
}

// Forbidden 无权限，HTTP 403
func Forbidden(msg string, err error) *AppError {
	return &AppError{Code: ErrCodeForbidden, HTTPStatus: http.StatusForbidden, Message: msg, Err: err}
}

// Conflict 资源冲突，HTTP 409
func Conflict(msg string, err error) *AppError {
	return &AppError{Code: ErrCodeConflict, HTTPStatus: http.StatusConflict, Message: msg, Err: err}
}

// K8sClientFail K8s 客户端创建失败，HTTP 502
func K8sClientFail(msg string, err error) *AppError {
	return &AppError{Code: ErrCodeK8sClient, HTTPStatus: http.StatusBadGateway, Message: msg, Err: err}
}

// K8sAPIFail K8s API 调用失败，HTTP 502
func K8sAPIFail(msg string, err error) *AppError {
	return &AppError{Code: ErrCodeK8sAPI, HTTPStatus: http.StatusBadGateway, Message: msg, Err: err}
}

// Internal 内部错误，HTTP 500
func Internal(msg string, err error) *AppError {
	return &AppError{Code: ErrCodeInternal, HTTPStatus: http.StatusInternalServerError, Message: msg, Err: err}
}
