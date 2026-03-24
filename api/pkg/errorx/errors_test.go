package errorx

import (
	"errors"
	"net/http"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name:     "无原始错误",
			err:      ErrMissingImageURL,
			expected: "缺少img参数",
		},
		{
			name:     "有原始错误",
			err:      Wrap(ErrImageDownload, errors.New("connection refused")),
			expected: "图片下载失败: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, 期望 %v", got, tt.expected)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("original error")
	err := Wrap(ErrImageDownload, cause)

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("Unwrap() = %v, 期望 %v", unwrapped, cause)
	}
}

func TestWrap(t *testing.T) {
	cause := errors.New("test error")
	wrapped := Wrap(ErrImageDownload, cause)

	if wrapped.Code != ErrImageDownload.Code {
		t.Errorf("Code = %v, 期望 %v", wrapped.Code, ErrImageDownload.Code)
	}

	if wrapped.HTTPStatus != ErrImageDownload.HTTPStatus {
		t.Errorf("HTTPStatus = %v, 期望 %v", wrapped.HTTPStatus, ErrImageDownload.HTTPStatus)
	}

	if wrapped.Cause != cause {
		t.Errorf("Cause = %v, 期望 %v", wrapped.Cause, cause)
	}
}

func TestWrapMessage(t *testing.T) {
	customMsg := "自定义错误消息"
	wrapped := WrapMessage(ErrImageDownload, customMsg)

	if wrapped.Message != customMsg {
		t.Errorf("Message = %v, 期望 %v", wrapped.Message, customMsg)
	}

	if wrapped.Code != ErrImageDownload.Code {
		t.Errorf("Code = %v, 期望 %v", wrapped.Code, ErrImageDownload.Code)
	}
}

func TestNew(t *testing.T) {
	code := "TEST_ERROR"
	message := "测试错误"
	httpStatus := http.StatusBadRequest

	err := New(code, message, httpStatus)

	if err.Code != code {
		t.Errorf("Code = %v, 期望 %v", err.Code, code)
	}

	if err.Message != message {
		t.Errorf("Message = %v, 期望 %v", err.Message, message)
	}

	if err.HTTPStatus != httpStatus {
		t.Errorf("HTTPStatus = %v, 期望 %v", err.HTTPStatus, httpStatus)
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        *AppError
		expectCode string
		expectStatus int
	}{
		{"ErrMissingImageURL", ErrMissingImageURL, "MISSING_IMAGE_URL", http.StatusBadRequest},
		{"ErrInvalidURL", ErrInvalidURL, "INVALID_URL", http.StatusBadRequest},
		{"ErrInvalidProtocol", ErrInvalidProtocol, "INVALID_PROTOCOL", http.StatusBadRequest},
		{"ErrSSRFAttack", ErrSSRFAttack, "SSRF_ATTACK", http.StatusForbidden},
		{"ErrImageTooLarge", ErrImageTooLarge, "IMAGE_TOO_LARGE", http.StatusRequestEntityTooLarge},
		{"ErrImageDownload", ErrImageDownload, "IMAGE_DOWNLOAD_FAILED", http.StatusBadGateway},
		{"ErrImageDecode", ErrImageDecode, "IMAGE_DECODE_FAILED", http.StatusUnsupportedMediaType},
		{"ErrForbidden", ErrForbidden, "FORBIDDEN", http.StatusForbidden},
		{"ErrRateLimitExceeded", ErrRateLimitExceeded, "RATE_LIMIT_EXCEEDED", http.StatusTooManyRequests},
		{"ErrInternalServer", ErrInternalServer, "INTERNAL_SERVER_ERROR", http.StatusInternalServerError},
		{"ErrTimeout", ErrTimeout, "TIMEOUT", http.StatusGatewayTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.expectCode {
				t.Errorf("Code = %v, 期望 %v", tt.err.Code, tt.expectCode)
			}

			if tt.err.HTTPStatus != tt.expectStatus {
				t.Errorf("HTTPStatus = %v, 期望 %v", tt.err.HTTPStatus, tt.expectStatus)
			}
		})
	}
}
