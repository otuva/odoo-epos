package transformer

import (
	"fmt"

	"github.com/xiaohao0576/odoo-epos/raster"
)

func init() {
	Transformers["receipt"] = func(input *raster.RasterImage) *raster.RasterImage {
		fmt.Printf("Original Image size: %dx%d\n", input.Width, input.Height)
		header := raster.NewRasterImageFromFile("header.png")
		table := input.WithCropRows(245, 350)
		img := input.WithDeleteRows(0, 430).WithDeleteRows(-219, -119)
		result := header.WithAppend(table).WithAppend(img)
		fmt.Printf("ReceiptTransformer: Result size: %dx%d\n", result.Width, result.Height)
		return result
	}
}
