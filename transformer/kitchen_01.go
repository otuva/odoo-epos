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
	SearchArea:  image.Rect(0, 200, 35, -1),
	Width:       35,
	Height:      40,
	BorderWidth: 4,
	WhitePixels: []image.Point{{0, 0}, {30, 8}, {31, 9}, {32, 10}, {33, 11}, {34, 12},
		{25, 20}, {26, 21}, {27, 22}, {28, 23}, {29, 24}, {30, 25}, {31, 26}, {32, 27}, {33, 28}, {34, 29}},
	BlackRatio: -0.05,
}

var ProductNumberOnePattern *raster.RasterPattern = &raster.RasterPattern{
	SearchArea: image.Rect(0, 200, 35, -1),
	Width:      35,
	Height:     40,
	BlackPixels: []image.Point{{13, 10}, {13, 11}, {13, 12}, {13, 13}, {13, 14}, {13, 15}, {13, 16}, {13, 17}, {13, 18}, {13, 19}, {13, 20},
		{13, 21}, {13, 22}, {13, 23}, {13, 24}, {13, 25}, {13, 26}, {13, 27}, {13, 28}, {13, 29},
		{14, 10}, {14, 11}, {14, 12}, {14, 13}, {14, 14}, {14, 15}, {14, 16}, {14, 17}, {14, 18}, {14, 19}, {14, 20},
		{14, 21}, {14, 22}, {14, 23}, {14, 24}, {14, 25}, {14, 26}, {14, 27}, {14, 28}, {14, 29}, {11, 10}, {12, 10}},
}
