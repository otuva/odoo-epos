package transformer

import "github.com/xiaohao0576/odoo-epos/raster"

type TransformerFunc func(input *raster.RasterImage) *raster.RasterImage

var Transformers = map[string]TransformerFunc{}

var Identity TransformerFunc = func(img *raster.RasterImage) *raster.RasterImage {
	return img
}

func init() {
	Transformers[""] = Identity
}
