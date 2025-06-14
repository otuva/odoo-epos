package raster

import (
	"fmt"
	"image"
)

type RasterPattern struct {
	width      int                 // 图像宽度
	height     int                 // 图像高度
	pointColor map[image.Point]int // 点颜色映射，0表示白色，1表示黑色
	blackRatio [2]float64          // 黑色像素点占比，两个浮点数
}

func NewRasterPattern(width, height int) *RasterPattern {
	if width <= 0 || height <= 0 {
		return nil // 无效的尺寸，返回 nil
	}
	return &RasterPattern{
		width:      width,
		height:     height,
		pointColor: make(map[image.Point]int), // 初始化点颜色映射
		blackRatio: [2]float64{0, 1},          // 默认黑色像素点比例为 0 到 1
	}
}

func (p *RasterPattern) SetSize(width, height int) {
	if width <= 0 || height <= 0 {
		return // 无效的尺寸，忽略
	}
	p.width = width
	p.height = height
	if p.pointColor == nil {
		p.pointColor = make(map[image.Point]int) // 初始化点颜色映射
	}
}

func (p *RasterPattern) Width() int {
	return p.width // 返回图案宽度
}

func (p *RasterPattern) Height() int {
	return p.height // 返回图案高度
}

func (p *RasterPattern) Size() image.Point {
	return image.Point{X: p.width, Y: p.height} // 返回图案尺寸
}

func (p *RasterPattern) SetBlackRatio(lower, upper float64) {
	if upper < 0 || lower < 0 || upper < lower {
		return // 无效的比例，忽略
	}
	p.blackRatio[0] = lower // 设置黑色像素点上限比例
	p.blackRatio[1] = upper // 设置黑色像素点下限比例
}

func (p *RasterPattern) GetBlackRatio() (lower, upper float64) {
	return p.blackRatio[0], p.blackRatio[1] // 返回黑色像素点比例
}

func (p *RasterPattern) AddBlackPoint(x, y int) {
	p.pointColor[image.Point{x, y}] = 1
}

func (p *RasterPattern) AddWhitePoint(x, y int) {
	p.pointColor[image.Point{x, y}] = 0
}

func (p *RasterPattern) AddBlackPoints(points []image.Point) {
	for _, pt := range points {
		p.AddBlackPoint(pt.X, pt.Y)
	}
}

func (p *RasterPattern) AddWhitePoints(points []image.Point) {
	for _, pt := range points {
		p.AddWhitePoint(pt.X, pt.Y)
	}
}

func (p *RasterPattern) DeletePoint(x, y int) {
	delete(p.pointColor, image.Point{x, y})
}

func (p *RasterPattern) DeletePoints(points []image.Point) {
	for _, pt := range points {
		p.DeletePoint(pt.X, pt.Y)
	}
}

func (p *RasterPattern) GetPointColor(x, y int) int {
	if color, exists := p.pointColor[image.Point{x, y}]; exists {
		return color // 返回点的颜色，0表示白色，1表示黑色
	}
	return -1 // 如果点不存在，返回 -1
}

func (p *RasterPattern) GetPointColors() map[image.Point]int {
	return p.pointColor // 返回所有点的颜色映射
}

func (p *RasterPattern) AddBlackRow(y int) {
	for x := 0; x < p.width; x++ {
		p.AddBlackPoint(x, y)
	}
}

func (p *RasterPattern) AddWhiteRow(y int) {
	for x := 0; x < p.width; x++ {
		p.AddWhitePoint(x, y)
	}
}

func (p *RasterPattern) AddBlackColumn(x int) {
	for y := 0; y < p.height; y++ {
		p.AddBlackPoint(x, y)
	}
}
func (p *RasterPattern) AddWhiteColumn(x int) {
	for y := 0; y < p.height; y++ {
		p.AddWhitePoint(x, y)
	}
}

func (p *RasterPattern) AddBlackRows(rows []int) {
	for y := range rows {
		p.AddBlackRow(y)
	}
}

func (p *RasterPattern) AddWhiteRows(rows []int) {
	for y := range rows {
		p.AddWhiteRow(y)
	}
}

func (p *RasterPattern) AddBlackColumns(cols []int) {
	for x := range cols {
		p.AddBlackColumn(x)
	}
}
func (p *RasterPattern) AddWhiteColumns(cols []int) {
	for x := range cols {
		p.AddWhiteColumn(x)
	}
}

func (p *RasterPattern) AddBlackArea(rect image.Rectangle) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			p.AddBlackPoint(x, y)
		}
	}
}

func (p *RasterPattern) AddWhiteArea(rect image.Rectangle) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			p.AddWhitePoint(x, y)
		}
	}
}

func (p *RasterPattern) DeleteArea(rect image.Rectangle) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			p.DeletePoint(x, y) // 删除指定区域的点
		}
	}
}

func (p *RasterPattern) AddBlackBorder(width int) {
	for y := 0; y < p.height; y++ {
		for x := 0; x < p.width; x++ {
			if x < width || x >= p.width-width || y < width || y >= p.height-width {
				p.AddBlackPoint(x, y) // 添加黑色边框点
			}
		}
	}
}

func (p *RasterPattern) AddWhiteBorder(width int) {
	for y := 0; y < p.height; y++ {
		for x := 0; x < p.width; x++ {
			if x < width || x >= p.width-width || y < width || y >= p.height-width {
				p.AddWhitePoint(x, y) // 添加白色边框点
			}
		}
	}
}

func (pattern *RasterPattern) IsMatch(img *RasterSubImage) bool {
	if img == nil || pattern == nil || pattern.width <= 0 || pattern.height <= 0 {
		return false // 无效的图像或图案，返回 false
	}

	if !img.Area.Size().Eq(pattern.Size()) {
		fmt.Println("Image size does not match pattern size:", img.Area.Size(), "vs", pattern.Size())

		return false // 如果裁剪区域大小与图案不匹配，返回 false
	}

	for point, color := range pattern.pointColor {
		if point.X < 0 {
			point.X = pattern.width + point.X // 支持负数索引
		}
		if point.Y < 0 {
			point.Y = pattern.height + point.Y // 支持负数索引
		}
		pixel := img.GetPointPixel(point)
		if pixel != color {
			return false // 如果颜色不匹配，返回 false
		}
	}

	// 黑色像素比例判断
	lower, upper := pattern.GetBlackRatio()
	if lower != 0 || upper != 0 {
		ratio := img.BlackRatio()
		if ratio > upper || ratio < lower {
			if ratio != 0 {
				fmt.Println("Black pixel ratio out of bounds:", ratio, "expected between", lower, "and", upper)
			}
			return false // 如果黑色像素比例不在范围内，返回 false
		}
	}

	return true
}

func (pattern *RasterPattern) IsMatchAt(img *RasterSubImage, x, y int) bool {
	subImage := img.Select(image.Rect(x, y, x+pattern.width, y+pattern.height))
	if subImage == nil {
		fmt.Println("Invalid sub-image at position:", x, y)
		return false // 如果裁剪区域无效，返回 false
	}
	return pattern.IsMatch(subImage) // 检查裁剪区域是否与图案匹
}

func (pattern *RasterPattern) SearchFirstMatch(img *RasterSubImage) *RasterSubImage {
	if img == nil || pattern == nil || pattern.width <= 0 || pattern.height <= 0 {
		return nil // 无效的图像或图案，返回 nil
	}

	imgWidth, imgHeight := img.Width(), img.Height()
	if imgWidth < pattern.width || imgHeight < pattern.height {
		return nil // 如果图像尺寸小于图案尺寸，返回 nil
	}

	for y := 0; y <= imgHeight-pattern.height; y++ {
		for x := 0; x <= imgWidth-pattern.width; x++ {
			if pattern.IsMatchAt(img, x, y) {
				return img.Select(image.Rect(x, y, x+pattern.width, y+pattern.height)) // 返回第一个匹配的裁剪区域
			}
		}
	}

	return nil // 如果没有找到匹配，返回 nil
}

func (pattern *RasterPattern) SearchAllMatches(img *RasterSubImage) []*RasterSubImage {
	if img == nil || pattern == nil || pattern.width <= 0 || pattern.height <= 0 {
		fmt.Printf("Invalid image or pattern, %s, %d, %d", img, pattern.width, pattern.height)
		return nil // 无效的图像或图案，返回 nil
	}
	fmt.Println("Searching for all matches of pattern at size", pattern.Size())

	imgWidth, imgHeight := img.Width(), img.Height()
	if imgWidth < pattern.width || imgHeight < pattern.height {
		return nil // 如果图像尺寸小于图案尺寸，返回 nil
	}

	var matches []*RasterSubImage

	for y := 0; y <= imgHeight-pattern.height; y++ {
		for x := 0; x <= imgWidth-pattern.width; x++ {
			if isMatched(img.GlobalPoint(x, y), matches) {
				fmt.Println("Position already matched:", x, y)
				continue // 如果当前点已经在匹配区域内，跳过
			}
			fmt.Println("Checking position:", x, y)
			if pattern.IsMatchAt(img, x, y) {
				matches = append(matches, img.Select(image.Rect(x, y, x+pattern.width, y+pattern.height))) // 添加匹配的裁剪区域
				x += pattern.width - 1                                                                     // 跳过已匹配的区域，避免重复匹配
			}
		}
	}

	return matches // 返回所有匹配的裁剪区域
}

func isMatched(point image.Point, matches []*RasterSubImage) bool {
	for _, img := range matches {
		if point.In(img.Area) {
			fmt.Println("Point is already matched in area:", img.Area)
			return true // 如果点在某个裁剪区域内，返回 true
		}
	}
	return false // 如果点不在任何裁剪区域内，返回 false
}

func (pattern *RasterPattern) MarkAllMatches(img *RasterSubImage) {
	matches := pattern.SearchAllMatches(img)
	fmt.Println("Found", len(matches), "matches for pattern at size", pattern.Size())
	for _, img := range matches {
		img.SetBorder()
	}
}
