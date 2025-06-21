package transformer

import (
	"fmt"
	"image"

	"github.com/xiaohao0576/odoo-epos/raster"
)

var NumberOCR *raster.RasterOCR

func init() {
	Transformers["reprint"] = func(input *raster.RasterImage) *raster.RasterImage {
		var orderNumber = getOrderNumber(input.SelectAll())
		input.SelectRows(0, 40).FillBlack()
		if orderNumber == nil {
			return input
		}
		var trackingNumber string
		for _, number := range orderNumber {
			char := NumberOCR.Recognize(number)
			trackingNumber += char
		}
		fmt.Printf("Tracking Number: %s\n", trackingNumber)
		return input
	}
}

func getOrderNumber(input *raster.RasterSubImage) []*raster.RasterSubImage {
	numberSignString := `iVBORw0KGgoAAAANSUhEUgAAABIAAAAXAQMAAAA8zY2nAAAABlBMVEUAAAD///+l2Z/dAAAAAWJL
R0QAiAUdSAAAAAlwSFlzAAALEwAACxMBAJqcGAAAAEtJREFUCNdj+Fd5gAGE/4Don0C6EIg/H2D4
8fEAQwMDAxA7MPwE8n8+B+LHEPojEAe6HgDKHmBY1OrA8Pn8AYbHQPwchPuBeD4YAwDcoDDpTcS/
xQAAAABJRU5ErkJggg==`
	numberSign := raster.NewSubImageFromBase64(numberSignString)
	subImg, _ := numberSign.MatchIn(input.Original.SelectAll())
	numberArea := image.Rect(subImg.Area.Max.X+3, subImg.Area.Min.Y, subImg.Area.Max.X+100, subImg.Area.Max.Y)
	numberSubImage := input.Original.Select(numberArea)
	numbers := numberSubImage.CutCharacters()
	return numbers
}

func init() {
	NumberOCR = &raster.RasterOCR{
		PngBase64: `iVBORw0KGgoAAAANSUhEUgAAAKAAAAAXAQMAAAC2ztajAAAABlBMVEX///8AAABVwtN+AAABVUlE
QVR4nCzOMY7UQBCF4WeVmCKwqlMHb+2NiRx2MDA3WRGS4dBCI21PRMgFkDgAZwD6CBxggxqNROwE
yUHLXtm76ae/Sg/iuO+gGU1Y0JzWyyQZYa78MXcTjmFFOa1fS5jQnYP/9/cjzmGpyruf3x7aAran
a56GI841JEMsHIlovOZhIEoNBZQaW8TD/S0NzuqJCDv2hl5xSz1YPdUwwCi9IR7gGED5TtQban/Y
yh2Vd3/5CbVtGDf8iFHZzXcFDBbtpXSdjN3v7g8aLa/nVVYn2baKKCXafi7pzUhucxDDQ/thLzXh
GDeUNJh2y14qwP4VqfoP/QHeAvWXXxBKcsrl5ScBPq576Y0kQzR4hNef10Ypl0xJ3HHAQLBRkypT
QNDgviPNVBI1jWiPuPrbccS2/YekrtMJYcZ1nU8TGIosAl2CQx23tYQMylTNCpkVzwEAAP//jld8
GsNkBdYAAAAASUVORK5CYII=`,
		CharAreas: map[string]image.Rectangle{
			"0": image.Rect(2, 0, 15, 23),
			"1": image.Rect(18, 0, 27, 23),
			"2": image.Rect(30, 0, 45, 23),
			"3": image.Rect(48, 0, 62, 23),
			"4": image.Rect(65, 0, 80, 23),
			"5": image.Rect(83, 0, 96, 23),
			"6": image.Rect(99, 0, 113, 23),
			"7": image.Rect(114, 0, 129, 23),
			"8": image.Rect(131, 0, 144, 23),
			"9": image.Rect(146, 0, 159, 23),
		},
	}
}
