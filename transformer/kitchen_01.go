package transformer

import (
	"image"

	"github.com/xiaohao0576/odoo-epos/raster"
)

func init() {
	Transformers["kitchen"] = func(input *raster.RasterImage) *raster.RasterImage {
		var ProductQtyPattern = raster.NewRasterPattern(35, 40)
		var SubImage = input.Select(image.Rect(0, 240, 35, input.Height))
		ProductQtyPattern.AddWhiteArea(image.Rect(0, 0, 35, 40))
		ProductQtyPattern.DeleteArea(image.Rect(5, 5, 25, 34))
		ProductQtyPattern.SetBlackRatio(0.05, 0.15)
		ProductQtyPattern.MarkAllMatches(SubImage)
		return input
	}
}
