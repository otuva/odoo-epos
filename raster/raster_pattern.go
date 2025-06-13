package raster

import (
	"fmt"
	"image"
)

type RasterPattern struct {
	NearBottom  bool            // 是否从底部开始搜索
	SearchArea  image.Rectangle // 搜索区域，默认为图像的全部区域
	Width       int             // 图像宽度
	Height      int             // 图像高度
	BorderWidth int             // 边框宽度
	BlackPixels []image.Point   // 黑色像素点的索引列表
	WhitePixels []image.Point   // 白色像素点的索引列表
	BlackRatio  float64         // 黑色像素点占比，正数表示上限，0表示不限制，负数表示下限
}

func (p *RasterPattern) AddBlackPixel(x, y int) {
	p.BlackPixels = append(p.BlackPixels, image.Point{X: x, Y: y})
}

func (p *RasterPattern) AddWhitePixel(x, y int) {
	p.WhitePixels = append(p.WhitePixels, image.Point{X: x, Y: y})
}

func (p *RasterPattern) AddBlackRows(rows []int) {
	for _, row := range rows {
		for x := 0; x < p.Width; x++ {
			p.BlackPixels = append(p.BlackPixels, image.Point{X: x, Y: row})
		}
	}
}

func (p *RasterPattern) AddWhiteRows(rows []int) {
	for _, row := range rows {
		for x := 0; x < p.Width; x++ {
			p.WhitePixels = append(p.WhitePixels, image.Point{X: x, Y: row})
		}
	}
}

func (p *RasterPattern) AddBlackColumns(columns []int) {
	for _, col := range columns {
		for y := 0; y < p.Height; y++ {
			p.BlackPixels = append(p.BlackPixels, image.Point{X: col, Y: y})
		}
	}
}

func (p *RasterPattern) AddWhiteColumns(columns []int) {
	for _, col := range columns {
		for y := 0; y < p.Height; y++ {
			p.WhitePixels = append(p.WhitePixels, image.Point{X: col, Y: y})
		}
	}
}

func (p *RasterPattern) AddBlackArea(rect image.Rectangle) {
	if rect.Empty() {
		return // 无效的矩形区域
	}
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			p.AddBlackPixel(x, y)
		}
	}
}

func (p *RasterPattern) AddWhiteArea(rect image.Rectangle) {
	if rect.Empty() {
		return // 无效的矩形区域
	}
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			p.AddWhitePixel(x, y)
		}
	}
}

func (p *RasterPattern) SetSearchArea(rect image.Rectangle) {
	if rect.Empty() {
		p.SearchArea = image.Rect(0, 0, p.Width, p.Height) // 默认搜索区域为图案的全部区域
	} else {
		p.SearchArea = rect
	}
}

// WithBorderPattern 在图像中绘制一个边框图案
// pattern: 要搜索的图案
func (img *RasterImage) WithBorderRect(rect image.Rectangle) *RasterImage {
	if img == nil || rect.Empty() {
		return img // 无效的图像或矩形区域，返回原图像
	}
	if rect.Max.X <= 0 || rect.Max.Y <= 0 || rect.Min.X >= img.Width || rect.Min.Y >= img.Height {
		return img // 矩形区域完全在图像外部，返回原图像
	}
	minX := rect.Min.X
	minY := rect.Min.Y
	maxX := rect.Max.X
	maxY := rect.Max.Y
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			if x == minX || x == maxX-1 || y == minY || y == maxY-1 {
				img.SetPixelBlack(x, y)
			}
		}
	}
	return img // 返回修改后的图像
}

// WithCropImage 在图像中裁剪出匹配的图案
func (img *RasterImage) WithCropPattern(pattern *RasterPattern) *RasterImage {
	x, y := img.SearchPattern(pattern)
	if x < 0 || y < 0 {
		return nil // 未找到匹配的图案，返回 nil
	}
	// 裁剪图像
	return img.WithCrop(x, y, pattern.Width, pattern.Height)
}

// WithErasePattern 返回擦除指定图案后的图像
// pattern: 要擦除的图案
// 如果图案无效（如宽度或高度小于等于0），则返回原图像
func (img *RasterImage) WithErasePattern(pattern *RasterPattern) *RasterImage {
	x, y := img.SearchPattern(pattern)
	if x < 0 || y < 0 {
		return img // 未找到匹配的图案，返回 nil
	}
	// 擦除区域
	return img.WithErase(x, y, pattern.Width, pattern.Height)
}

func (img *RasterImage) WithDeletePatternRows(pattern *RasterPattern) *RasterImage {
	x, y := img.SearchPattern(pattern)
	if x < 0 || y < 0 {
		return img // 未找到匹配的图案，返回原图像
	}
	// 删除匹配图案所在的行
	return img.WithDeleteRows(y, y+pattern.Height)
}

// WithBorderPatternAll 在图像中绘制所有匹配图案的边框
// pattern: 要搜索的图案
func (img *RasterImage) WithBorderPatternAll(pattern *RasterPattern) {
	if img == nil || pattern == nil || pattern.Width <= 0 || pattern.Height <= 0 {
		return // 无效的图像或图案，直接返回
	}

	matches := img.SearchPatternAll(pattern)
	for _, match := range matches {
		// 在每个匹配位置绘制边框
		img.WithBorderRect(image.Rect(match.X, match.Y, match.X+pattern.Width, match.Y+pattern.Height))
	}
}

// SearchPatternAll 在图像中搜索所有匹配的图案
// 返回所有匹配的图案位置列表
func (img *RasterImage) SearchPatternAll(pattern *RasterPattern) []image.Point {
	if img == nil || pattern == nil || pattern.Width <= 0 || pattern.Height <= 0 {
		return nil // 无效的图像或图案，返回 nil
	}

	var matches []image.Point
	area := pattern.SearchArea
	if area.Empty() {
		area = image.Rect(0, 0, img.Width, img.Height)
	}

	x1 := area.Min.X
	y1 := area.Min.Y
	x2 := area.Max.X
	y2 := area.Max.Y
	// 支持负数索引
	if x2 < 0 {
		x2 = img.Width + x2
	}
	if y2 < 0 {
		y2 = img.Height + y2
	}
	if x1 < 0 {
		x1 = img.Width + x1
	}
	if y1 < 0 {
		y1 = img.Height + y1
	}
	// 确保搜索区域在图像范围内
	if x2 > img.Width {
		x2 = img.Width
	}
	if y2 > img.Height {
		y2 = img.Height
	}
	area = image.Rect(x1, y1, x2, y2) // 更新

	minX := area.Min.X
	minY := area.Min.Y
	maxX := area.Max.X - pattern.Width  // 减去图案宽度
	maxY := area.Max.Y - pattern.Height // 减去图案高度

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			fmt.Printf("Checking pattern at (%d, %d)\n", x, y) // 输出检查位置
			if pattern.IsMatchAt(img, x, y) {
				matches = append(matches, image.Point{X: x, Y: y})
				x = x + pattern.Width - 1  // 跳过已匹配的区域
				y = y + pattern.Height - 1 // 跳过已匹配的区域
			}
		}
	}

	return matches // 返回所有匹配的图案位置列表
}

// SearchPattern 在图像中搜索匹配的图案
// 根据 SearchDirection 的值决定从顶部还是底部开始搜索
func (img *RasterImage) SearchPattern(pattern *RasterPattern) (int, int) {
	width := pattern.Width
	if width <= 0 {
		width = img.Width
	}
	height := pattern.Height
	if height <= 0 {
		height = img.Height
	}
	area := pattern.SearchArea
	if area.Empty() {
		area = image.Rect(0, 0, img.Width, img.Height)
	}
	// 后续都用 area 替换 pattern.SearchArea

	// 支持负数索引
	if area.Max.X < 0 {
		area.Max.X = img.Width + area.Max.X
	}
	if area.Max.Y < 0 {
		area.Max.Y = img.Height + area.Max.Y
	}
	if area.Min.X < 0 {
		area.Min.X = img.Width + area.Min.X
	}
	if area.Min.Y < 0 {
		area.Min.Y = img.Height + area.Min.Y
	}
	// 确保搜索区域在图像范围内
	if area.Max.X > img.Width {
		area.Max.X = img.Width
	}
	if area.Max.Y > img.Height {
		area.Max.Y = img.Height
	}

	if img == nil || img.Width < width || img.Height < height {
		return -1, -1
	}
	if pattern.NearBottom {
		return pattern.SearchFromBottom(img) // 如果指定了从底部开始搜索，调用 SearchFromBottom
	}
	return pattern.SearchFromTop(img) // 否则从顶部开始搜索
}

// SearchFromTop 从图像的左上角开始裁剪，逐行逐列检查，直到找到匹配的图案
// 返回匹配的图案在图像中的位置（x, y），如果未找到则返回 (-1, -1)
// 注意：这里假设图案的宽度和高度都大于0
// 如果图案的宽度或高度无效，或者图像无效或小于图案尺寸，则返回 (-1, -1)
// 只返回第一个匹配的图案位置
func (p *RasterPattern) SearchFromTop(img *RasterImage) (int, int) {
	area := p.SearchArea
	minX := area.Min.X
	minY := area.Min.Y
	maxX := area.Max.X - p.Width
	maxY := area.Max.Y - p.Height
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if p.IsMatchAt(img, x, y) {
				return x, y
			}
		}
	}
	return -1, -1
}

// SearchFromBottom 从图像的左下角开始裁剪，逐行逐列检查，直到找到匹配的图案
// 返回匹配的图案在图像中的位置（x, y），如果未找到则返回 (-1, -1)
// 注意：这里假设图案的宽度和高度都大于0
// 如果图案的宽度或高度无效，或者图像无效或小于图案尺寸，则返回 (-1, -1)
// 只返回第一个匹配的图案位置
func (p *RasterPattern) SearchFromBottom(img *RasterImage) (int, int) {
	area := p.SearchArea
	minX := area.Min.X
	maxX := area.Max.X - p.Width
	minY := area.Min.Y
	maxY := area.Max.Y - p.Height
	for y := maxY; y >= minY; y-- {
		for x := minX; x <= maxX; x++ {
			if p.IsMatchAt(img, x, y) {
				return x, y
			}
		}
	}
	return -1, -1
}

// IsMatchAt 在原图 img 的 (offsetX, offsetY) 区域判断是否与当前模式匹配（不创建裁剪实例）
func (p *RasterPattern) IsMatchAt(img *RasterImage, offsetX, offsetY int) bool {
	if img == nil {
		return false
	}

	// 尺寸判断
	if offsetX+p.Width > img.Width || offsetY+p.Height > img.Height {
		return false
	}

	if p.BorderWidth > 0 {
		// 取 (0,0) 位置的颜色作为边框颜色
		borderColor := img.GetPixel(offsetX, offsetY)
		// 检查边框区域颜色是否一致
		for y := 0; y < p.Height; y++ {
			for x := 0; x < p.Width; x++ {
				if x < p.BorderWidth || x >= p.Width-p.BorderWidth || y < p.BorderWidth || y >= p.Height-p.BorderWidth {
					if img.GetPixel(offsetX+x, offsetY+y) != borderColor {
						return false
					}
				}
			}
		}
	}

	// 黑色像素点判断
	for _, point := range p.BlackPixels {
		if point.X < 0 {
			point.X = p.Width + point.X // 支持负数索引
		}
		if point.Y < 0 {
			point.Y = p.Height + point.Y // 支持负数索引
		}
		pixel := img.GetPixel(offsetX+point.X, offsetY+point.Y)
		if pixel != 1 {
			return false
		}
	}

	// 白色像素点判断
	for _, point := range p.WhitePixels {
		if point.X < 0 {
			point.X = p.Width + point.X // 支持负数索引
		}
		if point.Y < 0 {
			point.Y = p.Height + point.Y // 支持负数索引
		}
		pixel := img.GetPixel(offsetX+point.X, offsetY+point.Y)
		if pixel != 0 {
			return false
		}
	}

	// 黑色像素比例判断
	if p.BlackRatio != 0 {
		blackCount := 0
		total := p.Width * p.Height

		for y := 0; y < p.Height; y++ {
			for x := 0; x < p.Width; x++ {
				if img.GetPixel(offsetX+x, offsetY+y) == 1 {
					blackCount++
				}
			}
		}
		blackRatio := float64(blackCount) / float64(total)

		if p.BlackRatio > 0 {
			// 正数表示上限
			if blackRatio > p.BlackRatio {
				return false
			}
		} else {
			// 负数表示下限
			if blackRatio < -p.BlackRatio {
				return false
			}
		}
	}

	return true
}
