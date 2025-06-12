package raster

import (
	"image"
)

type RasterPattern struct {
	NearBottom  bool            // 是否从底部开始搜索
	SearchArea  image.Rectangle // 搜索区域，默认为图像的全部区域
	Width       int             // 图像宽度
	Height      int             // 图像高度
	borderWidth int             // 边框宽度
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

// Search 在图像中搜索匹配的图案
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

	if p.borderWidth > 0 {
		// 取 (0,0) 位置的颜色作为边框颜色
		borderColor := img.GetPixel(offsetX, offsetY)
		// 检查边框区域颜色是否一致
		for y := 0; y < p.Height; y++ {
			for x := 0; x < p.Width; x++ {
				if x < p.borderWidth || x >= p.Width-p.borderWidth || y < p.borderWidth || y >= p.Height-p.borderWidth {
					if img.GetPixel(offsetX+x, offsetY+y) != borderColor {
						return false
					}
				}
			}
		}
	}

	// 黑色像素点判断
	for _, point := range p.BlackPixels {
		pixel := img.GetPixel(offsetX+point.X, offsetY+point.Y)
		if pixel != 1 {
			return false
		}
	}

	// 白色像素点判断
	for _, point := range p.WhitePixels {
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
