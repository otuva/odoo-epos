package raster

func (img *RasterImage) toGSV0() []byte {
	// 参数验证
	if img == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil {
		return nil
	}
	const GS = 0x1D // ESC/POS 命令前缀
	// 计算 xL 和 xH
	xL, xH := LowHighValue(img.Width / 8)
	// 计算 yL 和 yH
	yL, yH := LowHighValue(img.Height)

	result := append([]byte{GS, 'v', 48, 0, xL, xH, yL, yH}, img.Content...)

	return result
}

func (img *RasterImage) ToEscPosRasterCommand(maxHeight int) []byte {
	if maxHeight <= 0 || maxHeight > img.Height {
		return img.toGSV0()
	}
	result := make([]byte, 0, 100+len(img.Content))
	remainingHeight := img.Height
	offset := 0
	widthBytes := img.Width / 8
	const GS = 0x1D                       // ESC/POS 命令前缀
	xL, xH := LowHighValue(img.Width / 8) // 宽度不变，循环外计算
	for remainingHeight > 0 {
		currentHeight := min(remainingHeight, maxHeight)
		startPos := offset * widthBytes
		endPos := min(startPos+(currentHeight*widthBytes), len(img.Content))
		currentContent := img.Content[startPos:endPos]
		yL, yH := LowHighValue(currentHeight)
		command := append([]byte{GS, 'v', 48, 0, xL, xH, yL, yH}, currentContent...)
		result = append(result, command...)
		remainingHeight -= currentHeight
		offset += currentHeight
	}
	return result
}

func LowHighValue(value int) (low, high byte) {
	// 计算低位和高位字节
	low = byte(value & 0xFF)
	high = byte((value >> 8) & 0xFF)
	return
}
