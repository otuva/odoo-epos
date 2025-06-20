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
		for i, number := range orderNumber {
			number.PasteTo(input, 8+i*40, 8) // 每个数字占36像素高度
			for c, img := range NumberOCR.CharImages() {
				rate := img.MatchIn(number)
				fmt.Printf("Char: %s, Rate: %4f\n", c, rate)
			}
			fmt.Println("Next.......")
		}
		return input
	}
}

func getOrderNumber(input *raster.RasterSubImage) []*raster.RasterSubImage {
	numberSignPattern := raster.NewRasterPattern(24, 36)
	numberSignPattern.AddWhiteRows([]int{0, 1, 2, 3, 4, 5})
	numberSignPattern.AddWhiteRows([]int{35, 34, 33, 32, 31, 30})
	numberSignPattern.AddWhiteColumns([]int{0, 1, 2})
	numberSignPattern.AddWhiteColumns([]int{23, 22, 21})
	numberSignPattern.AddWhitePoints([]image.Point{
		{5, 10}, {7, 10}, {5, 8}, {7, 8}, {14, 7}, {14, 8}, {14, 9}, {19, 9},
		{6, 16}, {6, 17}, {6, 18}, {12, 16}, {12, 17}, {12, 18},
		{10, 24}, {10, 25}, {10, 26}, {10, 27}, {10, 28}, {18, 16}, {18, 17}, {18, 18},
	})
	numberSignPattern.AddBlackPoints([]image.Point{
		{10, 7}, {10, 8}, {10, 9}, {10, 10}, {10, 11}, {10, 12}, {10, 13}, {10, 14},
		{16, 8}, {16, 9}, {16, 10}, {16, 11}, {16, 12}, {16, 13}, {16, 14}, {16, 15},
		{6, 13}, {7, 13}, {8, 13}, {9, 13}, {11, 13}, {12, 13}, {13, 13}, {14, 13}, {15, 13},
		{17, 13}, {18, 13}, {19, 13}, {8, 17}, {8, 18}, {8, 19}, {8, 20}, {8, 21}, {8, 22}, {8, 23},
		{14, 18}, {14, 19}, {14, 20}, {14, 21}, {14, 22}, {14, 23}, {14, 24}, {14, 25}, {14, 26},
	})
	numberSign := numberSignPattern.SearchFirstMatch(input)
	if numberSign == nil {
		return nil
	}
	numberArea := image.Rect(numberSign.Area.Max.X, numberSign.Area.Min.Y, numberSign.Area.Max.X+100, numberSign.Area.Max.Y)
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
