package raster

import "strings"

func (img *RasterImage) AddMarginLeft(margin int) {
	if margin <= 0 {
		return // 如果边距小于等于0，则不添加边距
	}
	widthBytes := img.Width / 8
	marginBytes := margin / 8
	newWidth := img.Width + margin
	newContent := make([]byte, img.Height*(newWidth/8))
	for row := 0; row < img.Height; row++ {
		oldRowStart := row * widthBytes
		oldRowEnd := oldRowStart + widthBytes
		newRowStart := row * (newWidth / 8)
		// 前 marginBytes 字节自动为0（空白），直接拷贝原内容到新行后面
		copy(newContent[newRowStart+marginBytes:], img.Content[oldRowStart:oldRowEnd])
	}
	img.Width = newWidth
	img.Content = newContent
}

func (img *RasterImage) AddMarginRight(margin int) {
	if margin <= 0 {
		return // 如果边距小于等于0，则不添加边距
	}
	widthBytes := img.Width / 8
	newWidth := img.Width + margin
	newContent := make([]byte, img.Height*(newWidth/8))
	for row := 0; row < img.Height; row++ {
		oldRowStart := row * widthBytes
		oldRowEnd := oldRowStart + widthBytes
		newRowStart := row * (newWidth / 8)
		// 先拷贝原内容到新行前面，后面的空白字节自动为0（空白）
		copy(newContent[newRowStart:], img.Content[oldRowStart:oldRowEnd])
		// 右侧空白无需填充，默认就是0
	}
	img.Width = newWidth
	img.Content = newContent
}

func (img *RasterImage) AddMarginTop(margin int) {
	if margin < 0 {
		margin = 0
	}
	img.Height += margin
	newContent := make([]byte, len(img.Content)+margin*img.Width/8)
	// 将原内容拷贝到新内容的下方，前面部分自动为0（空白）
	copy(newContent[margin*img.Width/8:], img.Content)
	img.Content = newContent
}

func (img *RasterImage) AddMarginBottom(margin int) {
	if margin < 0 {
		margin = 0
	}
	img.Height += margin
	newContent := make([]byte, len(img.Content)+margin*img.Width/8)
	copy(newContent, img.Content)
	img.Content = newContent
}

func (img *RasterImage) AutoMarginLeft(paperWidth int) {
	if img.Width >= paperWidth {
		return // 如果图像宽度大于等于指定宽度，则不需要左边距
	}
	align := strings.ToLower(img.Align)
	margin := 0 // 默认左边距为0
	switch align {
	case "left":
		return // 左对齐不需要左边距
	case "right":
		margin = paperWidth - img.Width // 右对齐需要填充到指定宽度
	default:
		// 如果对齐方式不明确，默认使用居中对齐
		margin = (paperWidth - img.Width) / 2
	}
	margin = (margin / 8) * 8 // 确保左边距是8的倍数
	if margin > 0 {
		img.AddMarginLeft(margin)
	}
}

// AddMargin 添加左边距和下边距
func (img *RasterImage) AddMargin(marginLeft, marginBottom int) {
	if marginLeft < 0 {
		marginLeft = 0
	}
	if marginBottom < 0 {
		marginBottom = 0
	}
	img.AddMarginLeft(marginLeft)
	img.AddMarginBottom(marginBottom)
}
