package raster

import "time"

// WithDrawText 在图像上绘制文本
// text: 要绘制的文本内容
// x, y: 文本的起始位置坐标
// 返回值：返回一个新的RasterImage对象，包含绘制的文本
func (img *RasterImage) WithDrawText(text []rune, x, y int) *RasterImage {
	if img == nil || len(text) == 0 {
		return img // 如果图像为nil或文本为空，直接返回原图像
	}
	if img.Width <= 0 || img.Height <= 0 || img.Content == nil {
		return img // 如果图像无效，直接返回原图像
	}
	textImg := NewRasterImageFromText(text)
	if textImg == nil {
		return img
	}

	return img.WithPaste(textImg, x, y) // 将文本图像粘贴到指定位置
}

func NewRasterImageFromText(text []rune) *RasterImage {
	if len(text) == 0 {
		return nil // 如果文本为空，返回nil
	}
	width := len(text) * 16                 // 每个字符宽度为16像素
	height := 24                            // 每个字符高度为24像素
	content := make([]byte, height*width/8) // 初始化内容

	img := &RasterImage{
		Width:   width,
		Height:  height,
		Content: content,
	}

	for idx, r := range text {
		charBytes, ok := Fonts16x24[r]
		if !ok {
			continue // 跳过不存在的字符
		}
		for row := 0; row < 24; row++ {
			rowData := uint16(charBytes[row][0])<<8 | uint16(charBytes[row][1])
			for col := 0; col < 16; col++ {
				if rowData&(1<<(15-col)) != 0 {
					// 设置像素为黑色
					x := idx*16 + col
					y := row
					byteIndex := y*width/8 + x/8
					bitIndex := 7 - (x % 8)
					img.Content[byteIndex] |= 1 << bitIndex
				}
			}
		}
	}

	return img
}

// NewOrderTimeText 创建一个包含当前时间的文本图像
func NewOrderTimeText() *RasterImage {
	now := time.Now()
	timeStr := now.Format("2006-01-02 15:04:05")
	runes := []rune(timeStr)
	textImg := NewRasterImageFromText(runes)

	img := &RasterImage{
		Width:   512,
		Height:  60,
		Content: make([]byte, 60*512/8), // 初始化内容
		Align:   "center",               // 默认居中对齐
	}

	return img.WithPaste(textImg, 100, 15) // 将时间文本图像粘贴到指定位置
}
