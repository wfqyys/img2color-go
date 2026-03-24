package errorx

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// AppError 应用错误结构体
type AppError struct {
	Code       string // 错误代码
	Message    string // 错误消息
	HTTPStatus int    // HTTP状态码
	Cause      error  // 原始错误
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap 实现errors.Unwrap接口
func (e *AppError) Unwrap() error {
	return e.Cause
}

// 预定义错误类型
var (
	// 参数错误
	ErrMissingImageURL = &AppError{
		Code:       "MISSING_IMAGE_URL",
		Message:    "缺少img参数",
		HTTPStatus: http.StatusBadRequest,
	}

	// URL验证错误
	ErrInvalidURL = &AppError{
		Code:       "INVALID_URL",
		Message:    "无效的URL格式",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidProtocol = &AppError{
		Code:       "INVALID_PROTOCOL",
		Message:    "仅支持http和https协议",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrSSRFAttack = &AppError{
		Code:       "SSRF_ATTACK",
		Message:    "禁止访问内网地址",
		HTTPStatus: http.StatusForbidden,
	}

	// 图片错误
	ErrImageTooLarge = &AppError{
		Code:       "IMAGE_TOO_LARGE",
		Message:    "图片大小超过限制",
		HTTPStatus: http.StatusRequestEntityTooLarge,
	}

	ErrImageDownload = &AppError{
		Code:       "IMAGE_DOWNLOAD_FAILED",
		Message:    "图片下载失败",
		HTTPStatus: http.StatusBadGateway,
	}

	ErrImageDecode = &AppError{
		Code:       "IMAGE_DECODE_FAILED",
		Message:    "图片解码失败",
		HTTPStatus: http.StatusUnsupportedMediaType,
	}

	// 访问控制错误
	ErrForbidden = &AppError{
		Code:       "FORBIDDEN",
		Message:    "禁止访问",
		HTTPStatus: http.StatusForbidden,
	}

	ErrRateLimitExceeded = &AppError{
		Code:       "RATE_LIMIT_EXCEEDED",
		Message:    "请求过于频繁，请稍后再试",
		HTTPStatus: http.StatusTooManyRequests,
	}

	// 服务器错误
	ErrInternalServer = &AppError{
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "服务器内部错误",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrServiceUnavailable = &AppError{
		Code:       "SERVICE_UNAVAILABLE",
		Message:    "服务暂时不可用",
		HTTPStatus: http.StatusServiceUnavailable,
	}

	ErrTimeout = &AppError{
		Code:       "TIMEOUT",
		Message:    "请求超时",
		HTTPStatus: http.StatusGatewayTimeout,
	}
)

// ErrorResponse 错误响应结构体
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// writeErrorResponse 写入错误响应（内部函数）
func writeErrorResponse(w http.ResponseWriter, err error) {
	appErr, ok := err.(*AppError)
	if !ok {
		// 如果不是AppError，包装为内部服务器错误
		appErr = &AppError{
			Code:       ErrInternalServer.Code,
			Message:    err.Error(),
			HTTPStatus: http.StatusInternalServerError,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus)

	response := ErrorResponse{
		Code:    appErr.Code,
		Message: appErr.Message,
	}
	json.NewEncoder(w).Encode(response)
}

// WriteError 写入错误响应（公开函数，用于中间件和处理器）
var WriteError = writeErrorResponse

// Wrap 包装错误
func Wrap(err *AppError, cause error) *AppError {
	return &AppError{
		Code:       err.Code,
		Message:    err.Message,
		HTTPStatus: err.HTTPStatus,
		Cause:      cause,
	}
}

// WrapMessage 用自定义消息包装错误
func WrapMessage(err *AppError, message string) *AppError {
	return &AppError{
		Code:       err.Code,
		Message:    message,
		HTTPStatus: err.HTTPStatus,
	}
}

// New 创建新的应用错误
func New(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}
