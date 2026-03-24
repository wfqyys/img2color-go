package service

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestColorService_Extract(t *testing.T) {
	service := NewColorService()

	// 创建一个简单的测试图片
	img := createTestImage(100, 100, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	// 将图片编码为PNG
	var buf bytes.Buffer
	if err := encodePNG(&buf, img); err != nil {
		t.Fatalf("编码测试图片失败: %v", err)
	}

	// 提取颜色
	colorHex, err := service.Extract(buf.Bytes())
	if err != nil {
		t.Fatalf("颜色提取失败: %v", err)
	}

	// 验证颜色格式（应该是#RRGGBB格式）
	if len(colorHex) != 7 {
		t.Errorf("颜色格式不正确: %s (长度应为7)", colorHex)
	}

	if colorHex[0] != '#' {
		t.Errorf("颜色格式不正确: %s (应以#开头)", colorHex)
	}

	t.Logf("提取的颜色: %s", colorHex)
}

func TestColorService_ExtractDifferentColors(t *testing.T) {
	service := NewColorService()

	tests := []struct {
		name  string
		color color.RGBA
	}{
		{"红色", color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"绿色", color.RGBA{R: 0, G: 255, B: 0, A: 255}},
		{"蓝色", color.RGBA{R: 0, G: 0, B: 255, A: 255}},
		{"白色", color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"黑色", color.RGBA{R: 0, G: 0, B: 0, A: 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := createTestImage(100, 100, tt.color)

			var buf bytes.Buffer
			if err := encodePNG(&buf, img); err != nil {
				t.Fatalf("编码测试图片失败: %v", err)
			}

			colorHex, err := service.Extract(buf.Bytes())
			if err != nil {
				t.Fatalf("颜色提取失败: %v", err)
			}

			t.Logf("%s -> %s", tt.name, colorHex)
		})
	}
}

func TestColorService_decodeImage(t *testing.T) {
	service := NewColorService()

	// 测试无效的图片数据
	_, err := service.decodeImage([]byte("invalid image data"))
	if err == nil {
		t.Error("期望解码失败，但成功了")
	}
}

// createTestImage 创建测试图片
func createTestImage(width, height int, c color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

// encodePNG 编码PNG图片
func encodePNG(w *bytes.Buffer, img image.Image) error {
	return png.Encode(w, img)
}

