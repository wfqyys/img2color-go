package service

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"img2color-go/app/internal/pkg/errorx"
	"img2color-go/app/internal/pkg/httputil"
	"img2color-go/app/internal/pkg/logger"
)

// ImageService 图片服务
type ImageService struct {
	maxSize int64
}

// NewImageService 创建图片服务
func NewImageService(maxSize int64) *ImageService {
	return &ImageService{
		maxSize: maxSize,
	}
}

// Download 下载图片
func (s *ImageService) Download(ctx context.Context, url string) ([]byte, error) {
	// 创建请求
	req, err := httputil.CreateRequest(http.MethodGet, url)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrImageDownload, err)
	}

	// 使用带context的请求
	req = req.WithContext(ctx)

	// 发送请求
	client := httputil.GetHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("图片下载失败: %v", err)
		return nil, errorx.Wrap(errorx.ErrImageDownload, err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		logger.Error("图片下载失败，状态码: %d", resp.StatusCode)
		return nil, errorx.WrapMessage(errorx.ErrImageDownload, "图片下载失败，状态码: "+resp.Status)
	}

	// 检查Content-Length
	if resp.ContentLength > s.maxSize {
		logger.Warn("图片大小超过限制: %d > %d", resp.ContentLength, s.maxSize)
		return nil, errorx.ErrImageTooLarge
	}

	// 使用LimitReader限制读取大小
	limitedReader := io.LimitReader(resp.Body, s.maxSize+1)

	// 读取图片数据
	var buf bytes.Buffer
	n, err := io.Copy(&buf, limitedReader)
	if err != nil {
		logger.Error("读取图片数据失败: %v", err)
		return nil, errorx.Wrap(errorx.ErrImageDownload, err)
	}

	// 检查是否超过大小限制
	if n > s.maxSize {
		logger.Warn("图片大小超过限制: %d > %d", n, s.maxSize)
		return nil, errorx.ErrImageTooLarge
	}

	logger.Info("图片下载成功，大小: %d 字节", n)
	return buf.Bytes(), nil
}



