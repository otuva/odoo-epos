package raster

import (
	"image"
	"image/color"
)

type RasterSubImage struct {
	// The original image from which this sub-image is derived.
	Original *RasterImage
	// The rectangle defining the area of the original image that this sub-image represents.
	Area image.Rectangle
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

func (s *RasterSubImage) At(x, y int) color.Color {
	return s.Original.At(x+s.Area.Min.X, y+s.Area.Min.Y)
}

func (s *RasterSubImage) ColorModel() color.Model {
	return s.Original.ColorModel()
}

func (s *RasterSubImage) Size() (width, height int) {
	// Return the size of the sub-image, which is the width and height of the area defined in the RasterSubImage.
	return s.Area.Dx(), s.Area.Dy()
}

func (s *RasterSubImage) SetPixelBlack(x, y int) {
	if x < 0 {
		x = s.Area.Dx() + x // Adjust for negative coordinates
	}
	if y < 0 {
		y = s.Area.Dy() + y // Adjust for negative coordinates
	}
	// Set the pixel at (x, y) in the sub-image to black.
	if x >= s.Area.Dx() || y >= s.Area.Dy() {
		return // Out of bounds
	}
	// Calculate the corresponding pixel in the original image.
	originalX := x + s.Area.Min.X
	originalY := y + s.Area.Min.Y
	s.Original.SetPixelBlack(originalX, originalY)
}

func (s *RasterSubImage) SetPixelWhite(x, y int) {
	if x < 0 {
		x = s.Area.Dx() + x // Adjust for negative coordinates
	}
	if y < 0 {
		y = s.Area.Dy() + y // Adjust for negative coordinates
	}
	// Set the pixel at (x, y) in the sub-image to white.
	if x >= s.Area.Dx() || y >= s.Area.Dy() {
		return // Out of bounds
	}
	// Calculate the corresponding pixel in the original image.
	originalX := x + s.Area.Min.X
	originalY := y + s.Area.Min.Y
	s.Original.SetPixelWhite(originalX, originalY)
}

func (s *RasterSubImage) GetPixel(x, y int) int {
	if x < 0 {
		x = s.Area.Dx() + x // Adjust for negative coordinates
	}
	if y < 0 {
		y = s.Area.Dy() + y // Adjust for negative coordinates
	}
	// Get the pixel value at (x, y) in the sub-image.
	if x >= s.Area.Dx() || y >= s.Area.Dy() {
		return 0 // Out of bounds, return white (0)
	}
	// Calculate the corresponding pixel in the original image.
	originalX := x + s.Area.Min.X
	originalY := y + s.Area.Min.Y
	return s.Original.GetPixel(originalX, originalY)
}

// SubImage returns a sub-image of the original image defined by the area of this RasterSubImage.
func (rs *RasterImage) SubImage(area image.Rectangle) *RasterSubImage {
	if area.Min.X < 0 {
		area.Min.X += rs.Width // Adjust for negative coordinates
	}
	if area.Min.Y < 0 {
		area.Min.Y += rs.Height // Adjust for negative coordinates
	}
	if area.Max.X < 0 {
		area.Max.X += rs.Width // Adjust for negative coordinates
	}
	if area.Max.Y < 0 {
		area.Max.Y += rs.Height // Adjust for negative coordinates
	}
	// Clamp the area to the bounds of the original image using Intersect
	bounds := image.Rect(0, 0, rs.Width, rs.Height)
	area = area.Intersect(bounds)
	if area.Empty() {
		return nil // Invalid area
	}
	return NewRasterSubImage(rs, area)
}
