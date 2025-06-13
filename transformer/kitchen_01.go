package transformer

import (
	"fmt"
	"image"

	"github.com/xiaohao0576/odoo-epos/raster"
)

func init() {
	Transformers["kitchen"] = func(input *raster.RasterImage) *raster.RasterImage {
		fmt.Printf("Original Kitchen Image size: %dx%d\n", input.Width, input.Height)
		ProductQtyPattern.AddWhiteRows([]int{-8, -7, -6, -5, -4})
		input.WithBorderPatternAll(ProductQtyPattern)
		return input
	}
}

var ProductQtyPattern *raster.RasterPattern = &raster.RasterPattern{
	SearchArea: image.Rect(0, 200, 35, -1),
	Width:      35,
	Height:     40,
	WhitePixels: []image.Point{{0, 0}, {30, 8}, {31, 9}, {32, 10}, {33, 11}, {34, 12},
		{25, 20}, {26, 21}, {27, 22}, {28, 23}, {29, 24}, {30, 25}, {31, 26}, {32, 27}, {33, 28}, {34, 29}},
	BlackRatio: -0.05,
}
