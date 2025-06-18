package raster

import (
	"time"
)

// WithDrawText 在图像上绘制文本
// s: 要绘制的文本内容
// x, y: 文本的起始位置坐标
// 返回值：返回一个新的RasterImage对象，包含绘制的文本
func (img *RasterImage) WithDrawText(text string, x, y int) *RasterImage {
	textImg := NewRasterImageFromText(text)
	return img.WithPaste(textImg, x, y) // 将文本图像粘贴到指定位置
}

func (img *RasterImage) WithDrawInvertText(text string, x, y int) *RasterImage {
	textImg := NewRasterImageFromText(text)
	textImg.WithInvertPixel()           // 反转文本图像颜色
	return img.WithPaste(textImg, x, y) // 将文本图像粘贴到指定位置
}

func NewRasterImageFromText(s string) *RasterImage {
	text := []rune(s)
	font := Fonts16x24 // 使用16x24像素的字体
	charWidth := 16
	charHeight := 24
	width := len(text) * charWidth
	height := charHeight
	content := make([]byte, height*width/8)

	img := &RasterImage{
		Width:   width,
		Height:  height,
		Content: content,
	}

	for row := 0; row < charHeight; row++ {
		for i, r := range text {
			charBytes, ok := font[r]
			if !ok {
				charBytes = font[0] // 如果字符不存在，使用0值代替
			}
			// 每个字符一行2字节，拼接到目标内容
			offset := row*width/8 + i*2
			copy(img.Content[offset:offset+2], charBytes[row][:])
		}
	}
	return img
}

// NewOrderTimeText 创建一个包含当前时间的文本图像
func NewOrderTimeText() *RasterImage {
	now := time.Now()
	timeStr := now.Format("2006-01-02 15:04:05")
	textImg := NewRasterImageFromText(timeStr)

	img := &RasterImage{
		Width:   512,
		Height:  60,
		Content: make([]byte, 60*512/8), // 初始化内容
		Align:   "center",               // 默认居中对齐
	}

	return img.WithPaste(textImg, 190, 15) // 将时间文本图像粘贴到指定位置
}
