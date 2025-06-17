package transformer

import (
	"fmt"
	"image"

	"github.com/xiaohao0576/odoo-epos/raster"
)

func init() {
	Transformers["receipt"] = func(input *raster.RasterImage) *raster.RasterImage {
		fmt.Printf("Original Image size: %dx%d\n", input.Width, input.Height)
		header := raster.NewRasterImageFromFile("header.png")
		table := input.SelectRows(380, 490).Copy()
		img := input.WithDeleteRows(0, 580)
		img = withHideTime(img)
		result := header.WithAppend(table).WithAppend(img)
		fmt.Printf("ReceiptTransformer: Result size: %dx%d\n", result.Width, result.Height)
		return result
	}
}

func withHideTime(input *raster.RasterImage) *raster.RasterImage {
	timeArea := image.Rect(265, input.Height-37, 365, input.Height)
	dateArea := image.Rect(150, input.Height-37, 265, input.Height)
	input.Select(timeArea).FillWhite()
	dateImg := input.Select(dateArea).Cut()
	input = input.WithPaste(dateImg, 180, input.Height-37)
	return input

}
