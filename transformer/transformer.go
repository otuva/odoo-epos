package transformer

import "github.com/xiaohao0576/odoo-epos/raster"

type RasterTransformer interface {
	// Transform 将输入数据转换为输出数据
	Transform(input *raster.RasterImage) *raster.RasterImage
	// String 返回转换器的描述信息
	String() string
}

var Transformers = map[string]RasterTransformer{
	"": Identity,
}

var Identity = &IdentityTransformer{}

type IdentityTransformer struct{}

func (t *IdentityTransformer) Transform(input *raster.RasterImage) *raster.RasterImage {
	return input
}
func (t *IdentityTransformer) String() string {
	return "IdentityTransformer"
}
