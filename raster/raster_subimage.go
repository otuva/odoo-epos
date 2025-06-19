package raster

import (
	"image"
	"image/color"
	"iter"
)

type RasterSubImage struct {
	Original *RasterImage
	Area     image.Rectangle
}

// NewRasterSubImage creates a new RasterSubImage from the given original image and area.
func NewRasterSubImage(original *RasterImage, area image.Rectangle) *RasterSubImage {
	return &RasterSubImage{
		Original: original,
		Area:     area,
	}
}

func (s *RasterSubImage) Bounds() image.Rectangle {
	// Return the bounds of the sub-image, which is the area defined in the RasterSubImage.
	return s.Area
}

func (s *RasterSubImage) Width() int {
	// Return the width of the sub-image, which is the width of the area defined in the RasterSubImage.
	return s.Area.Dx()
}
func (s *RasterSubImage) Height() int {
	// Return the height of the sub-image, which is the height of the area defined in the RasterSubImage.
	return s.Area.Dy()
}

func (s *RasterSubImage) At(x, y int) color.Color {
	return s.Original.At(x+s.Area.Min.X, y+s.Area.Min.Y)
}

func (s *RasterSubImage) ColorModel() color.Model {
	return s.Original.ColorModel()
}

func (s *RasterSubImage) Size() (width, height int) {
	// Return the size of the sub-image, which is the width and height of the area defined in the RasterSubImage.
	return s.Width(), s.Height()
}

func (s *RasterSubImage) SetPixelBlack(x, y int) {
	originalX := x + s.Area.Min.X
	originalY := y + s.Area.Min.Y
	s.Original.SetPixelBlack(originalX, originalY)
}

func (s *RasterSubImage) SetPixelWhite(x, y int) {
	originalX := x + s.Area.Min.X
	originalY := y + s.Area.Min.Y
	s.Original.SetPixelWhite(originalX, originalY)
}

func (s *RasterSubImage) GetPixel(x, y int) int {
	originalX := x + s.Area.Min.X
	originalY := y + s.Area.Min.Y
	return s.Original.GetPixel(originalX, originalY)
}

func (s *RasterSubImage) GetPointPixel(point image.Point) int {
	return s.Original.GetPixel(point.X+s.Area.Min.X, point.Y+s.Area.Min.Y)
}

func (s *RasterSubImage) Copy() *RasterImage {
	if s.Area.Empty() {
		return nil // If the area is empty, return nil
	}
	rowBytes := (s.Width() + 7) / 8
	content := make([]byte, s.Height()*rowBytes)
	newImg := NewRasterImage(s.Width(), s.Height(), content)
	for y := s.Area.Min.Y; y < s.Area.Max.Y; y++ {
		for x := s.Area.Min.X; x < s.Area.Max.X; x++ {
			pixel := s.Original.GetPixel(x, y)
			point := image.Point{x, y}.Sub(s.Area.Min)
			newImg.SetPixel(point, pixel)
		}
	}
	return newImg
}

func (s *RasterSubImage) Cut() *RasterImage {
	img := s.Copy()
	s.FillWhite() // Fill the sub-image with white after cutting
	return img
}

func (s *RasterSubImage) PasteTo(target *RasterImage, x, y int) *RasterImage {
	return target.WithPaste(s.Original, x, y)

}

func (s *RasterSubImage) SetBorder() {
	// Set a border around the sub-image by setting the pixels at the edges to black.
	for y := s.Area.Min.Y; y < s.Area.Max.Y; y++ {
		for x := s.Area.Min.X; x < s.Area.Max.X; x++ {
			if x == s.Area.Min.X || x == s.Area.Max.X-1 || y == s.Area.Min.Y || y == s.Area.Max.Y-1 {
				s.Original.SetPixelBlack(x, y)
			}
		}
	}
}

func (s *RasterSubImage) FillBlack() {
	// Set all pixels in the sub-image to black.
	for y := s.Area.Min.Y; y < s.Area.Max.Y; y++ {
		for x := s.Area.Min.X; x < s.Area.Max.X; x++ {
			s.Original.SetPixelBlack(x, y)
		}
	}
}

func (s *RasterSubImage) FillWhite() {
	// Set all pixels in the sub-image to white.
	for y := s.Area.Min.Y; y < s.Area.Max.Y; y++ {
		for x := s.Area.Min.X; x < s.Area.Max.X; x++ {
			s.Original.SetPixelWhite(x, y)
		}
	}
}

func (s *RasterSubImage) InvertPixel() {
	// Invert the colors in the sub-image.
	for y := s.Area.Min.Y; y < s.Area.Max.Y; y++ {
		for x := s.Area.Min.X; x < s.Area.Max.X; x++ {
			if s.GetPixel(x, y) == 0 { // If pixel is white
				s.SetPixelBlack(x, y)
			} else { // If pixel is black
				s.SetPixelWhite(x, y)
			}
		}
	}
}

func (s *RasterSubImage) BlackRatio() float64 {
	// Calculate the ratio of black pixels in the sub-image.
	blackCount := 0
	totalCount := s.Width() * s.Height()
	for y := 0; y < s.Height(); y++ {
		for x := 0; x < s.Width(); x++ {
			if s.GetPixel(x, y) == 1 { // If pixel is black
				blackCount++
			}
		}
	}
	return float64(blackCount) / float64(totalCount)
}

func (s *RasterSubImage) GlobalX(x int) int {
	// Convert local x coordinate to global x coordinate in the original image.
	return x + s.Area.Min.X
}
func (s *RasterSubImage) GlobalY(y int) int {
	// Convert local y coordinate to global y coordinate in the original image.
	return y + s.Area.Min.Y
}

func (s *RasterSubImage) GlobalPoint(x, y int) image.Point {
	// Convert local point to global point in the original image.
	return image.Point{
		x + s.Area.Min.X,
		y + s.Area.Min.Y,
	}
}

func (s *RasterSubImage) Select(area image.Rectangle) *RasterSubImage {
	// Offset the area to the coordinates of the original image
	adjustedArea := area.Add(s.Area.Min)
	// Intersect with the parent area to ensure it's within bounds
	adjustedArea = adjustedArea.Intersect(s.Area)
	if adjustedArea.Empty() {
		return nil // Invalid area
	}
	return NewRasterSubImage(s.Original, adjustedArea)
}

func (s *RasterSubImage) Scan(pattern *RasterPattern) iter.Seq[*RasterSubImage] {
	return iter.Seq[*RasterSubImage](func(yield func(*RasterSubImage) bool) {
		if s == nil || pattern == nil || pattern.width <= 0 || pattern.height <= 0 {
			return // Invalid sub-image or pattern
		}
		imgWidth, imgHeight := s.Width(), s.Height()
		if imgWidth < pattern.width || imgHeight < pattern.height {
			return // Sub-image too small for the pattern
		}
		for y := 0; y <= imgHeight-pattern.height; y++ {
			for x := 0; x <= imgWidth-pattern.width; x++ {
				rect := image.Rect(x, y, x+pattern.width, y+pattern.height)
				if !yield(s.Select(rect)) {
					return
				}
			}
		}
	})
}

// SubImage returns a sub-image of the original image defined by the area of this RasterSubImage.
func (rs *RasterImage) Select(area image.Rectangle) *RasterSubImage {
	if area.Empty() {
		return nil // Invalid area
	}
	return NewRasterSubImage(rs, area)
}

func (rs *RasterImage) SelectAll() *RasterSubImage {
	area := image.Rect(0, 0, rs.Width, rs.Height)
	return rs.Select(area)
}

func (rs *RasterImage) SelectRows(y1, y2 int) *RasterSubImage {
	if y1 < 0 {
		y1 += rs.Height // Adjust for negative coordinates
	}
	if y2 < 0 {
		y2 += rs.Height // Adjust for negative coordinates
	}
	area := image.Rect(0, y1, rs.Width, y2)
	return rs.Select(area)
}

func (rs *RasterImage) SelectCols(x1, x2 int) *RasterSubImage {
	if x1 < 0 {
		x1 += rs.Width // Adjust for negative coordinates
	}
	if x2 < 0 {
		x2 += rs.Width // Adjust for negative coordinates
	}
	area := image.Rect(x1, 0, x2, rs.Height)
	return rs.Select(area)
}

func (rs *RasterSubImage) CutCharacters() []*RasterSubImage {
	width, height := rs.Width(), rs.Height()
	visited := make([][]bool, height)
	for i := range visited {
		visited[i] = make([]bool, width)
	}
	var chars []*RasterSubImage

	dx := []int{-1, -1, -1, 0, 0, 1, 1, 1}
	dy := []int{-1, 0, 1, -1, 1, -1, 0, 1}

	for y := range height {
		for x := range width {
			if visited[y][x] || rs.GetPixel(x, y) == 0 {
				continue
			}
			// BFS/DFS提取8连通区域
			minX, minY, maxX, maxY := x, y, x, y
			stack := []image.Point{{x, y}}
			visited[y][x] = true
			for len(stack) > 0 {
				pt := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				// 更新边界
				if pt.X < minX {
					minX = pt.X
				}
				if pt.Y < minY {
					minY = pt.Y
				}
				if pt.X > maxX {
					maxX = pt.X
				}
				if pt.Y > maxY {
					maxY = pt.Y
				}
				// 8方向
				for d := range 8 {
					nx, ny := pt.X+dx[d], pt.Y+dy[d]
					if nx >= 0 && nx < width && ny >= 0 && ny < height &&
						!visited[ny][nx] && rs.GetPixel(nx, ny) == 1 {
						visited[ny][nx] = true
						stack = append(stack, image.Point{nx, ny})
					}
				}
			}
			// 生成字符子图
			rect := image.Rect(minX, minY, maxX+1, maxY+1)
			sub := rs.Select(rect)
			if sub != nil {
				chars = append(chars, sub)
			}
		}
	}
	return chars
}
