package service

import (
	"bytes"
	"image"

	"img2color-go/app/core/pkg/errorx"
	"img2color-go/app/core/pkg/logger"

	"github.com/disintegration/imaging"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/nfnt/resize"
	"golang.org/x/image/webp"
)

// ColorService 颜色提取服务
type ColorService struct {
	resizeWidth uint
}

// NewColorService 创建颜色提取服务
func NewColorService() *ColorService {
	return &ColorService{
		resizeWidth: 50, // 缩放宽度为50像素
	}
}

// Extract 提取图片主色调
func (s *ColorService) Extract(imageData []byte) (string, error) {
	// 解码图片
	img, err := s.decodeImage(imageData)
	if err != nil {
		return "", err
	}

	// 缩放图片
	img = s.resizeImage(img)

	// 计算平均颜色
	color := s.calculateAverageColor(img)

	logger.Info("颜色提取成功: %s", color)
	return color, nil
}

// decodeImage 解码图片
func (s *ColorService) decodeImage(data []byte) (image.Image, error) {
	reader := bytes.NewReader(data)

	// 首先尝试使用imaging解码（支持JPEG, PNG, GIF, TIFF, BMP）
	img, err := imaging.Decode(reader)
	if err == nil {
		return img, nil
	}

	// 如果imaging解码失败，尝试WebP解码
	reader.Reset(data)
	img, err = webp.Decode(reader)
	if err == nil {
		logger.Info("使用WebP解码器成功")
		return img, nil
	}

	// 记录所有尝试的错误
	logger.Error("图片解码失败（尝试了JPEG, PNG, GIF, WebP等格式）: %v", err)
	return nil, errorx.Wrap(errorx.ErrImageDecode, err)
}

// resizeImage 缩放图片
func (s *ColorService) resizeImage(img image.Image) image.Image {
	return resize.Resize(s.resizeWidth, 0, img, resize.Lanczos3)
}

// calculateAverageColor 计算平均颜色
func (s *ColorService) calculateAverageColor(img image.Image) string {
	bounds := img.Bounds()
	var r, g, b uint32

	// 累加所有像素的RGB值
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r0, g0, b0, _ := c.RGBA()
			r += r0
			g += g0
			b += b0
		}
	}

	// 计算平均值
	totalPixels := uint32(bounds.Dx() * bounds.Dy())
	averageR := r / totalPixels
	averageG := g / totalPixels
	averageB := b / totalPixels

	// 转换为colorful.Color并获取十六进制值
	mainColor := colorful.Color{
		R: float64(averageR) / 0xFFFF,
		G: float64(averageG) / 0xFFFF,
		B: float64(averageB) / 0xFFFF,
	}

	return mainColor.Hex()
}




