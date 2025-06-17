package transformer

import (
	"image"

	"github.com/xiaohao0576/odoo-epos/raster"
)

func init() {
	Transformers["kitchen"] = func(input *raster.RasterImage) *raster.RasterImage {
		var header *raster.RasterSubImage
		if isKitchenCancelPattern(input) {
			input.SelectAll().InvertPixel()
			return input
		} else if isKitchenAddPattern(input) {
			header = input.Select(image.Rect(0, 0, input.Width, 185))
		} else if isKitchenDuplicataPattern(input) {
			header = input.Select(image.Rect(0, 0, input.Width, 240))
		} else {
			return input
		}
		var orderLines = searchKitchenOrderLines(input)
		input = input.WithCutline()
		for _, line := range orderLines {
			product := line.Copy().WithScaleY(2)
			input = input.WithAppend(header.Copy()).WithAppend(product).WithCutline()
		}
		return input
	}
}

func isKitchenCancelPattern(img *raster.RasterImage) bool {
	// Create a new raster pattern for cancel
	var pattern = raster.NewRasterPattern(512, 240)
	pattern.AddBlackPoints([]image.Point{
		{175, 208}, {168, 213}, {169, 226}, {176, 229}, {292, 229},
		{196, 209}, {191, 223}, {202, 223}, {226, 226}, {310, 218},
	})
	subImage := img.SelectAll()
	match := pattern.IsMatchAt(subImage, 0, 0)
	return match
}

func isKitchenAddPattern(img *raster.RasterImage) bool {
	var pattern = raster.NewRasterPattern(512, 240)
	pattern.AddBlackPoints([]image.Point{
		{224, 209}, {226, 213}, {224, 228}, {238, 227}, {248, 210},
		{248, 218}, {255, 218}, {247, 228}, {259, 229}, {265, 209}, {270, 225},
		{275, 215}, {283, 226}, {288, 209},
	})
	pattern.AddWhiteArea(image.Rect(0, 185, 210, 235))
	subImage := img.SelectAll()
	match := pattern.IsMatchAt(subImage, 0, 0)
	return match
}

func isKitchenDuplicataPattern(img *raster.RasterImage) bool {
	var pattern = raster.NewRasterPattern(512, 100)
	pattern.AddBlackPoints([]image.Point{
		{166, 60}, {166, 65}, {166, 81}, {170, 60}, {170, 81}, {178, 77}, {179, 70}, {187, 60},
		{187, 74}, {191, 81}, {194, 81}, {199, 79}, {201, 72}, {201, 63}, {211, 60}, {215, 72},
		{209, 80}, {217, 72}, {222, 69}, {220, 61}, {247, 62}, {247, 69}, {247, 80}, {263, 60},
		{255, 71}, {263, 81}, {270, 76}, {322, 75}, {303, 61},
	})
	subImage := img.SelectAll()
	match := pattern.IsMatchAt(subImage, 0, 0)
	return match
}

func searchKitchenOrderLines(input *raster.RasterImage) []*raster.RasterSubImage {
	var SubImage = input.Select(image.Rect(0, 240, 35, input.Height))
	var ProductQtyPattern = raster.NewRasterPattern(35, 40)
	ProductQtyPattern.AddWhiteArea(image.Rect(0, 0, 35, 40))
	ProductQtyPattern.DeleteArea(image.Rect(5, 5, 25, 34))
	ProductQtyPattern.SetBlackRatio(0.05, 0.15)
	matches := ProductQtyPattern.SearchAllMatches(SubImage)
	result := make([]*raster.RasterSubImage, 0, len(matches))
	for i, match := range matches {
		if i != len(matches)-1 {
			startY := match.Area.Min.Y
			endY := matches[i+1].Area.Min.Y
			result = append(result, input.SelectRows(startY, endY))
		} else {
			result = append(result, input.SelectRows(match.Area.Min.Y, input.Height))
		}
	}
	return result
}
