package raster

import "time"

func (img *RasterImage) DrawText(text []rune, x, y int) {

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
		Align:   "center",
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
