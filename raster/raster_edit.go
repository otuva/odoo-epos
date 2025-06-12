package raster

// WithCrop 返回裁剪后的图像
// x, y: 裁剪区域左上角坐标
// width, height: 裁剪区域的宽度和高度
// 如果裁剪区域超出原图范围，则返回nil
// 如果裁剪区域无效（如宽度或高度为0），也返回nil
func (img *RasterImage) WithCrop(x, y, width, height int) *RasterImage {
	if img == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil {
		return nil
	}

	// 支持负数索引
	if x < 0 {
		x = img.Width + x
	}
	if y < 0 {
		y = img.Height + y
	}

	// 检查参数有效性
	if x < 0 || y < 0 || width <= 0 || height <= 0 || x+width > img.Width || y+height > img.Height {
		return nil // 无效的裁剪参数
	}

	// 计算每行需要的字节数（向上取整到8的倍数）
	widthBytes := (width + 7) / 8
	croppedContent := make([]byte, height*widthBytes)

	for row := 0; row < height; row++ {
		srcBitStart := (y+row)*img.Width + x
		dstBitStart := row * width

		for col := 0; col < width; col++ {
			// 计算源像素位置
			srcIdx := srcBitStart + col
			srcByte := img.Content[srcIdx/8]
			srcBit := 7 - (srcIdx % 8)
			bit := (srcByte >> srcBit) & 1

			// 计算目标像素位置
			dstIdx := dstBitStart + col
			dstByteIdx := dstIdx / 8
			dstBitIdx := 7 - (dstIdx % 8)
			if bit != 0 {
				croppedContent[dstByteIdx] |= 1 << dstBitIdx
			}
			// 多余的bit自动为0（白色），无需处理
		}
	}

	return &RasterImage{
		Width:   width,
		Height:  height,
		Align:   img.Align,
		Content: croppedContent,
	}
}

// WithCropRows 返回裁剪后的图像，仅裁剪行
// startRow, endRow: 裁剪区域的起始行和结束行（包含）
// 如果裁剪区域超出原图范围，则返回nil
// 如果裁剪区域无效（如起始行大于结束行，或行数为0），也返回nil
func (img *RasterImage) WithCropRows(startRow, endRow int) *RasterImage {
	if img == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil {
		return nil
	}
	// 支持负数索引
	if startRow < 0 {
		startRow = img.Height + startRow
	}
	if endRow < 0 {
		endRow = img.Height + endRow
	}
	// 检查参数有效性
	if startRow < 0 || endRow < startRow || endRow >= img.Height {
		return nil // 无效的行范围
	}

	croppedHeight := endRow - startRow + 1
	widthBytes := img.Width / 8
	croppedContent := make([]byte, croppedHeight*widthBytes)
	for row := 0; row < croppedHeight; row++ {
		srcRowStart := (startRow + row) * widthBytes
		srcRowEnd := srcRowStart + widthBytes
		dstRowStart := row * widthBytes
		copy(croppedContent[dstRowStart:], img.Content[srcRowStart:srcRowEnd])
	}

	return &RasterImage{
		Width:   img.Width,
		Height:  croppedHeight,
		Align:   img.Align,
		Content: croppedContent,
	}
}

// WithDeleteRows 返回删除指定行后的图像
// startRow, endRow: 要删除的行范围（包含）
// 如果删除范围超出原图范围，则返回原图像
// 如果删除范围无效（如起始行大于结束行，或行数为0），也返回原图像
func (img *RasterImage) WithDeleteRows(startRow, endRow int) *RasterImage {
	if img == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil {
		return nil
	}
	// 支持负数索引
	if startRow < 0 {
		startRow = img.Height + startRow
	}
	if endRow < 0 {
		endRow = img.Height + endRow
	}
	if startRow < 0 {
		startRow = 0
	}
	if endRow >= img.Height {
		endRow = img.Height - 1
	}
	if startRow > endRow || startRow >= img.Height || endRow < 0 {
		return img // 无需删除
	}
	widthBytes := img.Width / 8
	newHeight := img.Height - (endRow - startRow + 1)
	if newHeight <= 0 {
		return nil // 全部删除
	}
	newContent := make([]byte, newHeight*widthBytes)
	dstRow := 0
	for row := 0; row < img.Height; row++ {
		if row < startRow || row > endRow {
			copy(
				newContent[dstRow*widthBytes:(dstRow+1)*widthBytes],
				img.Content[row*widthBytes:(row+1)*widthBytes],
			)
			dstRow++
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
		return nil
	}
	if other.Width != img.Width {
		return nil // 宽度不匹配，无法拼接
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
	// 支持负数索引
	if x < 0 {
		x = img.Width + x
	}
	if y < 0 {
		y = img.Height + y
	}

	if x < 0 || y < 0 || x >= img.Width || y >= img.Height {
		return img // 粘贴起点超出原图范围
	}

	// 自动裁剪other的宽高，防止越界
	maxWidth := img.Width - x
	maxHeight := img.Height - y
	pasteWidth := other.Width
	pasteHeight := other.Height
	if pasteWidth > maxWidth {
		pasteWidth = maxWidth
	}
	if pasteHeight > maxHeight {
		pasteHeight = maxHeight
	}

	newContent := make([]byte, len(img.Content))
	copy(newContent, img.Content)

	srcRowBytes := (other.Width + 7) / 8
	dstRowBytes := (img.Width + 7) / 8
	for row := 0; row < pasteHeight; row++ {
		srcRowStart := row * srcRowBytes
		dstRowStart := (y + row) * dstRowBytes
		for col := 0; col < pasteWidth; col++ {
			srcByteIdx := srcRowStart + col/8
			srcBitIdx := 7 - (col % 8)
			if (other.Content[srcByteIdx] & (1 << srcBitIdx)) != 0 {
				dstByteIdx := dstRowStart + (x+col)/8
				dstBitIdx := 7 - ((x + col) % 8)
				newContent[dstByteIdx] |= 1 << dstBitIdx
			}
		}
	}

	return &RasterImage{
		Width:   img.Width,
		Height:  img.Height,
		Align:   img.Align,
		Content: newContent,
	}
}

// WithErase 返回擦除后的图像
// x, y: 擦除区域的左上角坐标
// width, height: 擦除区域的宽度和高度
// 如果擦除区域超出原图像范围，则返回原图像
func (img *RasterImage) WithErase(x, y, width, height int) *RasterImage {
	if img == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil {
		return img
	}

	// 支持负数索引
	if x < 0 {
		x = img.Width + x
	}
	if y < 0 {
		y = img.Height + y
	}
	if x < 0 || y < 0 || width <= 0 || height <= 0 || x+width > img.Width || y+height > img.Height {
		return img
	}

	newContent := make([]byte, len(img.Content))
	copy(newContent, img.Content)
	widthBytes := (img.Width + 7) / 8

	for row := 0; row < height; row++ {
		rowStart := (y + row) * widthBytes
		for col := 0; col < width; col++ {
			byteIdx := rowStart + (x+col)/8
			bitIdx := 7 - ((x + col) % 8)
			newContent[byteIdx] &^= (1 << bitIdx)
		}
	}

	return &RasterImage{
		Width:   img.Width,
		Height:  img.Height,
		Align:   img.Align,
		Content: newContent,
	}
}

// WithBorder 返回添加边框后的图像
// borderWidth: 边框的宽度（单位为像素）
// 如果图像无效（如宽度或高度小于等于0，或内容为nil），则返回原图像
// 如果边框宽度小于等于0，则返回原图像
// 添加边框后，图像的宽度和高度保持不变
// 注意：边框会覆盖原图像的内容，原图像的内容将被边框覆盖
func (img *RasterImage) WithBorder(borderWidth int) *RasterImage {
	// 检查参数有效性
	if img == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil || borderWidth <= 0 {
		return img
	}

	newContent := make([]byte, len(img.Content))
	copy(newContent, img.Content)
	width := img.Width
	height := img.Height
	widthBytes := width / 8

	// 上下边框
	for row := 0; row < borderWidth; row++ {
		rowStart := row * widthBytes
		for i := 0; i < widthBytes; i++ {
			newContent[rowStart+i] = 0xFF // 全黑
		}
		rowStart = (height - 1 - row) * widthBytes
		for i := 0; i < widthBytes; i++ {
			newContent[rowStart+i] = 0xFF // 全黑
		}
	}
	// 左右边框
	for row := borderWidth; row < height-borderWidth; row++ {
		rowStart := row * widthBytes
		for col := 0; col < borderWidth; col++ {
			byteIdx := rowStart + col/8
			bitIdx := 7 - (col % 8)
			newContent[byteIdx] |= 1 << bitIdx // 左边
			byteIdx = rowStart + (width-1-col)/8
			bitIdx = 7 - ((width - 1 - col) % 8)
			newContent[byteIdx] |= 1 << bitIdx // 右边
		}
	}

	return &RasterImage{
		Width:   img.Width,
		Height:  img.Height,
		Align:   img.Align,
		Content: newContent,
	}
}

// shiftContent 返回内容向左或向右移动指定的位数
// shift: 正数表示向右移动，负数表示向左移动
// 如果移动位数为0，则返回原图像
func (img *RasterImage) shiftContent(shift int) *RasterImage {
	if img == nil || img.Content == nil || img.Width <= 0 || img.Height <= 0 {
		return img
	}
	if shift == 0 {
		return img
	}

	widthBytes := img.Width / 8
	newContent := make([]byte, len(img.Content))

	if shift > 0 { // 向右移动
		for row := 0; row < img.Height; row++ {
			srcRowStart := row * widthBytes
			dstRowStart := srcRowStart + shift/8
			for col := 0; col < widthBytes-shift/8; col++ {
				newContent[dstRowStart+col] = img.Content[srcRowStart+col]
			}
			// 填充左侧空白为0
			for col := 0; col < shift/8; col++ {
				newContent[srcRowStart+col] = 0
			}
		}
	} else { // 向左移动
		for row := 0; row < img.Height; row++ {
			srcRowStart := row * widthBytes
			dstRowStart := srcRowStart - (-shift / 8)
			for col := -shift / 8; col < widthBytes; col++ {
				newContent[dstRowStart+col] = img.Content[srcRowStart+col]
			}
			// 填充右侧空白为0
			for col := widthBytes + shift/8; col < widthBytes; col++ {
				newContent[srcRowStart+col] = 0
			}
		}
	}

	return &RasterImage{
		Width:   img.Width,
		Height:  img.Height,
		Align:   img.Align,
		Content: newContent,
	}
}

func (img *RasterImage) WithShiftLeft(shift int) *RasterImage {
	if shift <= 0 {
		return img // 如果移动位数小于等于0，则返回原图像
	}
	return img.shiftContent(-shift) // 向左移动
}

func (img *RasterImage) WithShiftRight(shift int) *RasterImage {
	if shift <= 0 {
		return img // 如果移动位数小于等于0，则返回原图像
	}
	return img.shiftContent(shift) // 向右移动
}
