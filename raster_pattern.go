package main

import "image"

type RasterPattern struct {
	NearBottom    bool          // 是否从底部开始搜索
	Width         int           // 图像宽度
	Height        int           // 图像高度
	IsWhiteBorder bool          // 是否有白色边框
	BlackPixels   []image.Point // 黑色像素点的索引列表
	WhitePixels   []image.Point // 白色像素点的索引列表
	BlackRatio    float64       // 黑色像素点占比
	WhiteRatio    float64       // 白色像素点占比
}

func (p *RasterPattern) AddBlackPixel(x, y int) {
	p.BlackPixels = append(p.BlackPixels, image.Point{X: x, Y: y})
}

func (p *RasterPattern) AddWhitePixel(x, y int) {
	p.WhitePixels = append(p.WhitePixels, image.Point{X: x, Y: y})
}

// CropImage 在图像中裁剪出匹配的图案
func (p *RasterPattern) CropImage(img *RasterImage) *RasterImage {
	x, y := p.Search(img)
	if x < 0 || y < 0 {
		return nil // 未找到匹配的图案，返回 nil
	}
	// 裁剪图像
	return img.WithCrop(x, y, p.Width, p.Height)
}

// Search 在图像中搜索匹配的图案
// 根据 SearchDirection 的值决定从顶部还是底部开始搜索
func (p *RasterPattern) Search(img *RasterImage) (int, int) {
	if p.Width <= 0 {
		p.Width = img.Width
	}
	if p.Height <= 0 {
		p.Height = img.Height
	}
	if img == nil || img.Width < p.Width || img.Height < p.Height {
		return -1, -1
	}
	if p.NearBottom {
		return p.SearchFromBottom(img) // 如果指定了从底部开始搜索，调用 SearchFromBottom
	}
	return p.SearchFromTop(img) // 否则从顶部开始搜索
}

// SearchFromTop 从图像的左上角开始裁剪，逐行逐列检查，直到找到匹配的图案
// 返回匹配的图案在图像中的位置（x, y），如果未找到则返回 (-1, -1)
// 注意：这里假设图案的宽度和高度都大于0
// 如果图案的宽度或高度无效，或者图像无效或小于图案尺寸，则返回 (-1, -1)
// 只返回第一个匹配的图案位置
func (p *RasterPattern) SearchFromTop(img *RasterImage) (int, int) {
	maxX := img.Width - p.Width
	maxY := img.Height - p.Height
	for y := 0; y <= maxY; y++ {
		for x := 0; x <= maxX; x++ {
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
// 这种方法适用于需要从底部开始搜索的场景，例如打印机的底部对齐
// 这种方法可以用于处理一些特殊的图像模式或打印机对齐方式
// 例如，当图案位于图像的底部时，可以使用此方法来快速找到匹配位置
func (p *RasterPattern) SearchFromBottom(img *RasterImage) (int, int) {
	maxX := img.Width - p.Width
	maxY := img.Height - p.Height
	for y := maxY; y >= 0; y-- {
		for x := 0; x <= maxX; x++ {
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
	// 白色边框判断
	if p.IsWhiteBorder {
		// 上下边
		for x := 0; x < p.Width; x++ {
			if img.GetPixel(offsetX+x, offsetY) != 0 { // 上
				return false
			}
			if img.GetPixel(offsetX+x, offsetY+p.Height-1) != 0 { // 下
				return false
			}
		}
		// 左右边
		for y := 1; y < p.Height-1; y++ {
			if img.GetPixel(offsetX, offsetY+y) != 0 { // 左
				return false
			}
			if img.GetPixel(offsetX+p.Width-1, offsetY+y) != 0 { // 右
				return false
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

	// 黑白比例判断
	if p.WhiteRatio > 0 || p.BlackRatio > 0 {
		blackCount := 0
		total := p.Width * p.Height
		maxBlack := int(float64(total) * (1 - p.WhiteRatio))
		minBlack := int(float64(total) * p.BlackRatio)
		remain := total

		for y := 0; y < p.Height; y++ {
			for x := 0; x < p.Width; x++ {
				remain--
				if img.GetPixel(offsetX+x, offsetY+y) == 1 {
					blackCount++
					if blackCount > maxBlack {
						return false // 黑色像素太多，提前终止
					}
				}
				// 如果剩余像素全是黑色也不够 minBlack，提前终止
				if blackCount+remain < minBlack {
					return false // 黑色像素太少，提前终止
				}
			}
		}
		blackRatio := float64(blackCount) / float64(total)
		whiteRatio := 1 - blackRatio
		if blackRatio < p.BlackRatio || whiteRatio < p.WhiteRatio {
			return false
		}
	}

	return true
}
