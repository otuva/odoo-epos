package transformer

import (
	"fmt"
	"image"

	"github.com/xiaohao0576/odoo-epos/raster"
)

func init() {
	Transformers["reprint"] = func(input *raster.RasterImage) *raster.RasterImage {
		var orderNumber = getOrderNumber(input.SelectAll())
		input.SelectRows(0, 40).FillBlack()
		if orderNumber == nil {
			return input
		}
		for i, number := range orderNumber {
			input = input.WithPaste(number.Copy(), i*30+8, 8)
			for c, pattern := range numberPatterns {
				if matched := pattern.SearchFirstMatch(number); matched != nil {
					fmt.Println("找到数字:", matched.Area, "匹配的数字:", c)
					break // 找到第一个匹配的数字就停止
				}
			}
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

var numberPatterns = map[string]*raster.RasterPattern{}

func init() {
	// 预定义数字的图案
	numberPatterns["0"] = raster.NewRasterPattern(14, 23)
	numberPatterns["0"].AddWhiteArea(image.Rect(4, 5, 9, 18))
	numberPatterns["0"].AddBlackArea(image.Rect(0, 5, 2, 18))
	numberPatterns["0"].AddBlackArea(image.Rect(11, 4, 13, 19))

	numberPatterns["1"] = raster.NewRasterPattern(9, 23)
	numberPatterns["1"].AddWhiteArea(image.Rect(0, 7, 6, 23))
	numberPatterns["1"].AddBlackArea(image.Rect(6, 1, 9, 23))
	numberPatterns["1"].AddBlackPoints([]image.Point{{2, 3}, {3, 3}, {4, 3}, {5, 3}})
	numberPatterns["1"].AddWhitePoints([]image.Point{{0, 0}, {1, 0}, {0, 1}})

	numberPatterns["2"] = raster.NewRasterPattern(14, 23)
	numberPatterns["2"].AddWhiteArea(image.Rect(0, 8, 3, 17))
	numberPatterns["2"].AddWhiteArea(image.Rect(0, 8, 6, 14))
	numberPatterns["2"].AddWhiteArea(image.Rect(5, 4, 9, 10))
	numberPatterns["2"].AddWhiteArea(image.Rect(10, 15, 14, 19))
	numberPatterns["2"].AddBlackArea(image.Rect(2, 21, 13, 23))

	numberPatterns["3"] = raster.NewRasterPattern(14, 23)
	numberPatterns["3"].AddWhiteArea(image.Rect(0, 7, 9, 8))
	numberPatterns["3"].AddWhiteArea(image.Rect(0, 7, 3, 16))
	numberPatterns["3"].AddWhiteArea(image.Rect(0, 14, 9, 16))
	numberPatterns["3"].AddWhiteArea(image.Rect(5, 14, 9, 19))
	numberPatterns["3"].AddWhiteArea(image.Rect(5, 4, 9, 8))
	numberPatterns["3"].AddBlackArea(image.Rect(6, 10, 10, 12))
	numberPatterns["3"].AddBlackArea(image.Rect(11, 3, 13, 9))

	numberPatterns["4"] = raster.NewRasterPattern(15, 23)
	numberPatterns["4"].AddWhiteArea(image.Rect(0, 0, 7, 3))
	numberPatterns["4"].AddWhiteArea(image.Rect(0, 0, 3, 9))
	numberPatterns["4"].AddWhiteArea(image.Rect(6, 11, 9, 14))
	numberPatterns["4"].AddBlackArea(image.Rect(10, 2, 13, 23))
	numberPatterns["4"].AddBlackArea(image.Rect(4, 16, 13, 18))

	numberPatterns["5"] = raster.NewRasterPattern(13, 23)
	numberPatterns["5"].AddWhiteArea(image.Rect(4, 4, 13, 7))
	numberPatterns["5"].AddWhiteArea(image.Rect(0, 13, 8, 16))
	numberPatterns["5"].AddWhiteArea(image.Rect(4, 12, 7, 19))

	numberPatterns["6"] = raster.NewRasterPattern(14, 23)
	numberPatterns["6"].AddWhiteArea(image.Rect(4, 13, 9, 17))
	numberPatterns["6"].AddWhiteArea(image.Rect(6, 5, 14, 7))
	numberPatterns["6"].AddWhiteArea(image.Rect(0, 0, 3, 2))

	numberPatterns["7"] = raster.NewRasterPattern(15, 23)
	numberPatterns["7"].AddWhiteArea(image.Rect(0, 4, 3, 23))
	numberPatterns["7"].AddWhiteArea(image.Rect(0, 4, 7, 11))

	numberPatterns["8"] = raster.NewRasterPattern(13, 23)
	numberPatterns["8"].AddWhiteArea(image.Rect(5, 4, 8, 8))
	numberPatterns["8"].AddWhiteArea(image.Rect(4, 5, 9, 8))
	numberPatterns["8"].AddWhiteArea(image.Rect(4, 15, 9, 19))
	numberPatterns["8"].AddBlackArea(image.Rect(4, 10, 10, 12))

	numberPatterns["9"] = raster.NewRasterPattern(13, 23)
	numberPatterns["9"].AddWhiteArea(image.Rect(4, 5, 9, 11))
	numberPatterns["9"].AddWhiteArea(image.Rect(0, 17, 8, 18))
	numberPatterns["9"].AddWhiteArea(image.Rect(0, 15, 1, 23))
	numberPatterns["9"].AddBlackArea(image.Rect(12, 5, 13, 16))
	numberPatterns["9"].AddBlackArea(image.Rect(0, 13, 2, 21))

}
