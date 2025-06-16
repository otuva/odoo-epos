package raster

import "bytes"

// Each byte has 7 bits set to 1 and 1 bit set to 0 (random position)
var CUTLINE = []byte{
	0xDF, 0x7F, 0xFB, 0xF7, 0xBF, 0xEF, 0xFD, 0xFE, 0xF7, 0xDF, 0x7F, 0xFB, 0xEF, 0xBF, 0xFE, 0xFD,
	0xFB, 0xF7, 0xDF, 0x7F, 0xBF, 0xEF, 0xFD, 0xFE, 0xF7, 0xDF, 0x7F, 0xFB, 0xEF, 0xBF, 0xFE, 0xFD,
	0xFE, 0xBF, 0xDF, 0x7F, 0xF7, 0xFB, 0xEF, 0xFD, 0x7F, 0xDF, 0xFB, 0xF7, 0xBF, 0xEF, 0xFE, 0xFD,
	0xF7, 0xFB, 0xDF, 0x7F, 0xBF, 0xEF, 0xFD, 0xFE, 0xFB, 0xF7, 0xDF, 0x7F, 0xEF, 0xBF, 0xFE, 0xFD,
	0xFD, 0xEF, 0xBF, 0x7F, 0xF7, 0xDF, 0xFB, 0xFE, 0xBF, 0xEF, 0x7F, 0xDF, 0xFB, 0xF7, 0xFE, 0xFD,
	0xEF, 0xBF, 0xDF, 0x7F, 0xF7, 0xFB, 0xFE, 0xFD, 0x7F, 0xDF, 0xFB, 0xF7, 0xBF, 0xEF, 0xFD, 0xFE,
	0xF7, 0xFB, 0xDF, 0x7F, 0xBF, 0xEF, 0xFE, 0xFD, 0xFB, 0xF7, 0xDF, 0x7F, 0xEF, 0xBF, 0xFE, 0xFD,
	0xFE, 0xBF, 0xDF, 0x7F, 0xF7, 0xFB, 0xEF, 0xFD, 0x7F, 0xDF, 0xFB, 0xF7, 0xBF, 0xEF, 0xFD, 0xFE,
}

// 在图像底部添加一行切割线
func (img *RasterImage) WithCutline() *RasterImage {
	if img == nil || img.Content == nil || img.Width <= 0 || img.Height <= 0 {
		return nil
	}

	// 确保宽度是8的倍数
	width := img.Width
	if width%8 != 0 {
		width = (width/8 + 1) * 8
	}

	bytesPerRow := width / 8
	cutline := make([]byte, bytesPerRow)
	copy(cutline, CUTLINE)

	// 复制原内容
	newContent := make([]byte, len(img.Content))
	copy(newContent, img.Content)
	newContent = append(newContent, cutline...)

	return &RasterImage{
		Width:   width,
		Height:  img.Height + 1,
		Content: newContent,
	}
}

func isCutline(line []byte) bool {
	if len(line) > len(CUTLINE) {
		return false
	}
	return bytes.Equal(line, CUTLINE[:len(line)])
}

// CutPages 将图像内容按切割线分割成多个页面
// 返回一个包含多个页面的切片，每个页面都是一个新的 *RasterImage 对象
func (img *RasterImage) CutPages() []*RasterImage {
	if img == nil || img.Content == nil || img.Height <= 0 {
		return nil
	}

	bytesPerRow := img.Width / 8
	var pages []*RasterImage
	var currContent []byte
	var currHeight int
	foundCutline := false

	// 忽略最后一行的 cutline
	lastLineIsCutline := false
	if img.Height > 0 {
		lastLine := img.Content[(img.Height-1)*bytesPerRow : img.Height*bytesPerRow]
		lastLineIsCutline = isCutline(lastLine)
	}

	// 只遍历到倒数第二行（如果最后一行是 cutline）
	limit := img.Height
	if lastLineIsCutline {
		limit = img.Height - 1
	}

	for i := 0; i < limit; i++ {
		line := img.Content[i*bytesPerRow : (i+1)*bytesPerRow]
		if isCutline(line) {
			foundCutline = true
			// 当前页结束，保存
			if currHeight > 0 {
				pages = append(pages, &RasterImage{
					Width:   img.Width,
					Height:  currHeight,
					Content: currContent,
				})
			}
			// 开始新页
			currContent = nil
			currHeight = 0
		} else {
			currContent = append(currContent, line...)
			currHeight++
		}
	}
	// 收尾：最后一页
	if currHeight > 0 {
		pages = append(pages, &RasterImage{
			Width:   img.Width,
			Height:  currHeight,
			Content: currContent,
		})
	}

	// 如果没有 cutline，返回原图
	if !foundCutline {
		return []*RasterImage{img}
	}
	return pages
}
