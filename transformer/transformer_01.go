package transformer

import (
	"image"

	"github.com/xiaohao0576/odoo-epos/raster"
)

func init() {
	Transformers["receipt"] = func(input *raster.RasterImage) *raster.RasterImage {
		header := raster.NewRasterImageFromFile("header.png")
		table := input.SelectRows(380, 490).Copy()
		img := input.WithDeleteRows(0, 580)
		img = withHideTime(img)
		result := header.WithAppend(table).WithAppend(img)
		return result
	}
}

func withHideTime(input *raster.RasterImage) *raster.RasterImage {
	timeArea := image.Rect(263, input.Height-37, 365, input.Height)
	input.Select(timeArea).FillWhite()
	offset := 40 // 偏移量，表示从左边开始的像素数
	for y := input.Height - 37; y < input.Height; y++ {
		for x := input.Width; x > offset; x-- {
			pixel := input.GetPixel(x-offset, y)
			input.SetPixel(image.Point{x, y}, pixel)
		}
	}
	return input

}
