package raster

import "image"

// WithCrop 返回裁剪后的图像
func (img *RasterImage) WithCrop(area image.Rectangle) *RasterImage {
	if area.Empty() {
		return nil
	}
	return img.Select(area).Copy()
}

// WithDeleteRows 返回删除指定行后的图像
// startRow, endRow: 要删除的行范围（包含）
// 如果删除范围超出原图范围，则返回原图像
// 如果删除范围无效（如起始行大于结束行，或行数为0），也返回原图像
func (img *RasterImage) WithDeleteRows(startRow, endRow int) *RasterImage {
	if startRow < 0 {
		startRow += img.Height // 支持负数索引
	}
	if endRow < 0 {
		endRow += img.Height // 支持负数索引
	}
	newHeight := img.Height - (endRow - startRow + 1)
	if newHeight <= 0 || startRow < 0 || endRow < startRow || endRow >= img.Height {
		return img // 无效的删除范围，返回原图像
	}
	newContent := make([]byte, newHeight*img.Width/8)
	for y := 0; y < img.Height; y++ {
		if y < startRow || y > endRow {
			rowContent := img.GetRow(y)
			copy(newContent[(y-(endRow-startRow+1))*img.Width/8:], rowContent)
		}
	}
	return &RasterImage{
		Width:   img.Width,
		Height:  newHeight,
		Align:   img.Align,
		Content: newContent,
	}
}

// WithAppend 返回拼接后的图像
// other: 要拼接的另一个图像
// 如果原图像或其他图像无效（如宽度或高度小于等于0，或内容为nil），则返回nil
// 如果其他图像的宽度与原图像不匹配，则返回nil
// 拼接后的图像高度为原图像高度加上其他图像高度
// 拼接后的图像宽度与原图像相同
func (img *RasterImage) WithAppend(other *RasterImage) *RasterImage {
	if img == nil || other == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil {
		return img
	}
	if other.Width != img.Width {
		return img // 宽度不匹配，无法拼接
	}

	newHeight := img.Height + other.Height
	newContent := make([]byte, newHeight*img.Width/8)
	copy(newContent, img.Content)
	copy(newContent[img.Height*img.Width/8:], other.Content)

	return &RasterImage{
		Width:   img.Width,
		Height:  newHeight,
		Align:   img.Align,
		Content: newContent,
	}
}

// WithPaste 返回粘贴后的图像
// other: 要粘贴的另一个图像
// x, y: 粘贴位置的左上角坐标
// 如果粘贴位置超出原图像范围，则返回原图像
func (img *RasterImage) WithPaste(other *RasterImage, x, y int) *RasterImage {
	// 检查参数有效性
	if img == nil || other == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil || other.Content == nil {
		return img
	}

	for dy := range other.Height {
		for dx := range other.Width {
			color := other.GetPixel(dx, dy)
			point := image.Point{x + dx, y + dy}
			img.SetPixel(point, color)
		}
	}

	return img
}

// WithScaleY 按整数倍 scale 放大图像的高度
func (img *RasterImage) WithScaleY(scale int) *RasterImage {
	if img == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil || scale <= 0 {
		return img
	}

	newHeight := img.Height * scale
	newWidth := img.Width
	widthBytes := (newWidth + 7) / 8
	newContent := make([]byte, widthBytes*newHeight)

	for y := 0; y < img.Height; y++ {
		rowContent := img.GetRow(y)
		for s := 0; s < scale; s++ {
			start := (y*scale + s) * widthBytes
			copy(newContent[start:start+widthBytes], rowContent)
		}
	}

	return &RasterImage{
		Width:   newWidth,
		Height:  newHeight,
		Align:   img.Align,
		Content: newContent,
	}
}
