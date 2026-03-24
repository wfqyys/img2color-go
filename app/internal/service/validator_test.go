package service

import (
	"testing"

	"img2color-go/app/internal/pkg/errorx"
)

func TestValidator_ValidateURL(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		url       string
		expectErr *errorx.AppError
	}{
		// 正常URL
		{
			name:      "有效的HTTP URL",
			url:       "http://example.com/image.jpg",
			expectErr: nil,
		},
		{
			name:      "有效的HTTPS URL",
			url:       "https://example.com/image.jpg",
			expectErr: nil,
		},

		// 协议验证
		{
			name:      "无效的file协议",
			url:       "file:///etc/passwd",
			expectErr: errorx.ErrInvalidProtocol,
		},
		{
			name:      "无效的ftp协议",
			url:       "ftp://example.com/file",
			expectErr: errorx.ErrInvalidProtocol,
		},

		// SSRF防护 - IPv4私有地址
		{
			name:      "SSRF防护 - 127.0.0.1",
			url:       "http://127.0.0.1/image.jpg",
			expectErr: errorx.ErrSSRFAttack,
		},
		{
			name:      "SSRF防护 - 10.0.0.1",
			url:       "http://10.0.0.1/image.jpg",
			expectErr: errorx.ErrSSRFAttack,
		},
		{
			name:      "SSRF防护 - 172.16.0.1",
			url:       "http://172.16.0.1/image.jpg",
			expectErr: errorx.ErrSSRFAttack,
		},
		{
			name:      "SSRF防护 - 192.168.1.1",
			url:       "http://192.168.1.1/image.jpg",
			expectErr: errorx.ErrSSRFAttack,
		},
		{
			name:      "SSRF防护 - 0.0.0.0",
			url:       "http://0.0.0.0/image.jpg",
			expectErr: errorx.ErrSSRFAttack,
		},

		// SSRF防护 - IPv6地址
		{
			name:      "SSRF防护 - ::1 (IPv6回环)",
			url:       "http://[::1]/image.jpg",
			expectErr: errorx.ErrSSRFAttack,
		},

		// SSRF防护 - 特殊主机名
		{
			name:      "SSRF防护 - localhost",
			url:       "http://localhost/image.jpg",
			expectErr: errorx.ErrSSRFAttack,
		},
		{
			name:      "SSRF防护 - localhost.localdomain",
			url:       "http://localhost.localdomain/image.jpg",
			expectErr: errorx.ErrSSRFAttack,
		},
		{
			name:      "SSRF防护 - .local域名",
			url:       "http://test.local/image.jpg",
			expectErr: errorx.ErrSSRFAttack,
		},
		{
			name:      "SSRF防护 - .internal域名",
			url:       "http://test.internal/image.jpg",
			expectErr: errorx.ErrSSRFAttack,
		},

		// 无效URL
		{
			name:      "无效的URL格式",
			url:       "not-a-url",
			expectErr: errorx.ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateURL(tt.url)

			if tt.expectErr == nil {
				if err != nil {
					t.Errorf("期望成功，但得到错误: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("期望错误 %v，但得到nil", tt.expectErr)
					return
				}

				appErr, ok := err.(*errorx.AppError)
				if !ok {
					t.Errorf("错误类型不是AppError: %v", err)
					return
				}

				if appErr.Code != tt.expectErr.Code {
					t.Errorf("错误代码不匹配: 期望 %s, 得到 %s", tt.expectErr.Code, appErr.Code)
				}
			}
		})
	}
}

func TestValidator_isPrivateIP(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name       string
		ip         string
		expectBool bool
	}{
		{"回环地址 127.0.0.1", "127.0.0.1", true},
		{"私有地址 10.0.0.1", "10.0.0.1", true},
		{"私有地址 172.16.0.1", "172.16.0.1", true},
		{"私有地址 192.168.1.1", "192.168.1.1", true},
		{"链路本地 169.254.1.1", "169.254.1.1", true},
		{"公网地址 8.8.8.8", "8.8.8.8", false},
		{"公网地址 1.1.1.1", "1.1.1.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 注意：isPrivateIP是私有方法，这里通过ValidateURL间接测试
			// 或者我们可以直接测试isPrivateAddress方法
			result := validator.isPrivateAddress(tt.ip)
			if result != tt.expectBool {
				t.Errorf("isPrivateAddress(%s) = %v, 期望 %v", tt.ip, result, tt.expectBool)
			}
		})
	}
}



