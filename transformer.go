package main

type RasterTransformer interface {
	// Transform 将输入数据转换为输出数据
	Transform(input *RasterImage) *RasterImage
	// String 返回转换器的描述信息
	String() string
}

var Transformers = map[string]RasterTransformer{
	"": &IdentityTransformer{},
}

type IdentityTransformer struct{}

func (t *IdentityTransformer) Transform(input *RasterImage) *RasterImage {
	return input
}
func (t *IdentityTransformer) String() string {
	return "IdentityTransformer"
}
