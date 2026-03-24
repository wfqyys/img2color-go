package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"img2color-go/app/internal/pkg/errorx"
	"img2color-go/app/internal/pkg/logger"
	"img2color-go/app/internal/service"
)

// Handler API处理器
type Handler struct {
	extractor       *service.ColorExtractor
	downloadTimeout time.Duration
}

// NewHandler 创建处理器
func NewHandler(extractor *service.ColorExtractor, downloadTimeout int) *Handler {
	return &Handler{
		extractor:       extractor,
		downloadTimeout: time.Duration(downloadTimeout) * time.Second,
	}
}

// ServeHTTP 处理HTTP请求
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 只处理GET请求
	if r.Method != http.MethodGet {
		errorx.WriteError(w, errorx.New("METHOD_NOT_ALLOWED", "仅支持GET请求", http.StatusMethodNotAllowed))
		return
	}

	// 获取图片URL参数
	imgURL := r.URL.Query().Get("img")
	if imgURL == "" {
		errorx.WriteError(w, errorx.ErrMissingImageURL)
		return
	}

	logger.Info("处理请求: %s", imgURL)

	// 提取颜色
	color, err := h.extractor.ExtractWithTimeout(imgURL, h.downloadTimeout)
	if err != nil {
		errorx.WriteError(w, err)
		return
	}

	// 返回响应
	response := map[string]string{
		"RGB": color,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("响应编码失败: %v", err)
	}
}

// HandleImageColor 处理图片颜色请求（兼容旧接口）
func (h *Handler) HandleImageColor(w http.ResponseWriter, r *http.Request) {
	h.ServeHTTP(w, r)
}


