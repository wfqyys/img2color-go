package service

import (
	"net"
	"net/url"
	"strings"

	"img2color-go/api/pkg/errorx"
)

// Validator URL验证器
type Validator struct {
	allowedProtocols []string
}

// NewValidator 创建验证器
func NewValidator() *Validator {
	return &Validator{
		allowedProtocols: []string{"http", "https"},
	}
}

// ValidateURL 验证URL合法性，防止SSRF攻击
func (v *Validator) ValidateURL(rawURL string) error {
	// 解析URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return errorx.Wrap(errorx.ErrInvalidURL, err)
	}

	// 验证协议
	if !v.isAllowedProtocol(parsedURL.Scheme) {
		return errorx.ErrInvalidProtocol
	}

	// 验证主机名
	if parsedURL.Hostname() == "" {
		return errorx.ErrInvalidURL
	}

	// 检查是否为内网地址
	if v.isPrivateAddress(parsedURL.Hostname()) {
		return errorx.ErrSSRFAttack
	}

	return nil
}

// isAllowedProtocol 检查协议是否允许
func (v *Validator) isAllowedProtocol(protocol string) bool {
	protocol = strings.ToLower(protocol)
	for _, allowed := range v.allowedProtocols {
		if protocol == allowed {
			return true
		}
	}
	return false
}

// isPrivateAddress 检查是否为私有地址（内网地址）
func (v *Validator) isPrivateAddress(hostname string) bool {
	// 尝试解析为IP地址
	ip := net.ParseIP(hostname)
	if ip != nil {
		return v.isPrivateIP(ip)
	}

	// 如果不是IP地址，尝试DNS解析
	// 注意：这里我们只检查hostname，不进行实际的DNS解析
	// 因为DNS解析可能会被攻击者利用进行DNS重绑定攻击

	// 检查常见的内网主机名
	privateHostnames := []string{
		"localhost",
		"localhost.localdomain",
		"ip6-localhost",
		"ip6-loopback",
	}

	hostnameLower := strings.ToLower(hostname)
	for _, private := range privateHostnames {
		if hostnameLower == private {
			return true
		}
	}

	// 检查以.local结尾的主机名（mDNS）
	if strings.HasSuffix(hostnameLower, ".local") {
		return true
	}

	// 检查以.internal结尾的主机名
	if strings.HasSuffix(hostnameLower, ".internal") {
		return true
	}

	return false
}

// isPrivateIP 检查IP是否为私有IP
func (v *Validator) isPrivateIP(ip net.IP) bool {
	// 检查是否为回环地址
	if ip.IsLoopback() {
		return true
	}

	// 检查是否为私有网络地址
	if ip.IsPrivate() {
		return true
	}

	// 检查是否为未指定地址（0.0.0.0）
	if ip.IsUnspecified() {
		return true
	}

	// 检查是否为链路本地地址
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// 额外检查一些特殊的私有地址段
	// 这些地址段在某些情况下可能不被IsPrivate()识别

	// IPv4私有地址段
	privateIPv4Networks := []string{
		"0.0.0.0/8",       // 当前网络
		"10.0.0.0/8",      // 私有网络
		"127.0.0.0/8",     // 回环地址
		"169.254.0.0/16",  // 链路本地
		"172.16.0.0/12",   // 私有网络
		"192.0.0.0/24",    // IANA保留
		"192.0.2.0/24",    // TEST-NET-1
		"192.88.99.0/24",  // IPv6到IPv4中继
		"192.168.0.0/16",  // 私有网络
		"198.18.0.0/15",   // 网络基准测试
		"198.51.100.0/24", // TEST-NET-2
		"203.0.113.0/24",  // TEST-NET-3
		"224.0.0.0/4",     // 多播
		"240.0.0.0/4",     // 保留
		"255.255.255.255/32", // 广播
	}

	// IPv6私有地址段
	privateIPv6Networks := []string{
		"::1/128",         // 回环地址
		"::/128",          // 未指定地址
		"::ffff:0:0/96",   // IPv4映射地址
		"fe80::/10",       // 链路本地
		"fc00::/7",        // 唯一本地地址
		"ff00::/8",        // 多播
	}

	// 检查IPv4私有地址
	if ip.To4() != nil {
		for _, network := range privateIPv4Networks {
			_, ipNet, err := net.ParseCIDR(network)
			if err == nil && ipNet.Contains(ip) {
				return true
			}
		}
	} else {
		// 检查IPv6私有地址
		for _, network := range privateIPv6Networks {
			_, ipNet, err := net.ParseCIDR(network)
			if err == nil && ipNet.Contains(ip) {
				return true
			}
		}
	}

	return false
}
