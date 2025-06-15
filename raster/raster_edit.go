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

	for row := range height {
		srcBitStart := (y+row)*img.Width + x
		dstBitStart := row * width

		for col := range width {
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

	for dy := range other.Height {
		for dx := range other.Width {
			color := other.GetPixel(dx, dy)
			img.setPixel(x+dx, y+dy, color)
		}
	}

	return img
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
	if x < 0 || y < 0 || width <= 0 || height <= 0 {
		return img
	}

	if x+width > img.Width {
		width = img.Width - x // 调整宽度，防止越界
	}

	if y+height > img.Height {
		height = img.Height - y // 调整高度，防止越界
	}

	newContent := make([]byte, len(img.Content))
	copy(newContent, img.Content)
	widthBytes := (img.Width + 7) / 8

	for row := range height {
		rowStart := (y + row) * widthBytes
		for col := range width {
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
