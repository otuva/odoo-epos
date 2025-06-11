package transformer

import (
	"fmt"

	"github.com/xiaohao0576/odoo-epos/raster"
)

func init() {
	Transformers["receipt"] = &ReceiptTransformer{
		Header: "收据头部",
		Footer: "收据底部",
	}
}

type ReceiptTransformer struct {
	Header string `json:"header"` // 收据头部
	Footer string `json:"footer"` // 收据底部
}

func (t *ReceiptTransformer) Transform(input *raster.RasterImage) *raster.RasterImage {
	fmt.Printf("Original Image size: %dx%d\n", input.Width, input.Height)
	header := raster.NewRasterImageFromFile("header.png")
	table := input.WithCrop(0, 380, input.Width, 120)
	img := input.WithDeleteRows(0, 570).WithDeleteRows(-219, -119)
	result := header.WithAppend(table).WithAppend(img)
	fmt.Printf("ReceiptTransformer: Result size: %dx%d\n", result.Width, result.Height)
	return result

}

func (t *ReceiptTransformer) String() string {
	return "ReceiptTransformer: " + t.Header + " | " + t.Footer
}
