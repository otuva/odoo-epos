package transformer

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"time"

	"github.com/xiaohao0576/odoo-epos/raster"
)

func init() {
	var cancelImg = getCancelImg()
	Transformers["kitchen"] = func(input *raster.RasterImage) *raster.RasterImage {
		var header *raster.RasterImage
		if isKitchenCancelPattern(input) {
			input = input.WithPaste(cancelImg, 0, 205)
			return input.AddMarginBottom(120)
		} else if isKitchenAddPattern(input) {
			header = input.Select(image.Rect(0, 100, input.Width, 210)).Copy()
		} else if isKitchenDuplicataPattern(input) {
			return nil // 取消打印
		} else {
			return input.AddMarginBottom(120)
		}
		var orderLines = searchKitchenOrderLines(input)
		header = header.AddMarginBottom(1)
		header.WithDrawText(time.Now().Format("01/02 15:04"), 0, 50)
		for _, line := range orderLines {
			input = input.WithAppend(header).WithCutline()
			product := line.Copy().AddMarginBottom(20)
			input = input.WithAppend(product)
		}
		return input.AddMarginBottom(120)
	}
}

func isKitchenCancelPattern(img *raster.RasterImage) bool {
	// Create a new raster pattern for cancel
	var pattern = raster.NewRasterPattern(512, 280)
	pattern.AddBlackPoints([]image.Point{
		{147, 235}, {139, 242}, {140, 258}, {177, 237}, {176, 256},
		{215, 259}, {227, 250}, {257, 249}, {282, 240}, {305, 253},
	})
	subImage := img.SelectAll()
	match := pattern.IsMatchAt(subImage, 0, 0)
	return match
}

func isKitchenAddPattern(img *raster.RasterImage) bool {
	var pattern = raster.NewRasterPattern(512, 280)
	pattern.AddBlackPoints([]image.Point{
		{214, 236}, {214, 243}, {214, 261}, {231, 259}, {234, 261},
		{245, 236}, {245, 249}, {245, 262}, {253, 236}, {253, 264}, {275, 260},
		{283, 239}, {291, 258}, {294, 259},
	})
	pattern.AddWhiteArea(image.Rect(0, 215, 195, 270))
	subImage := img.SelectAll()
	match := pattern.IsMatchAt(subImage, 0, 0)
	return match
}

func isKitchenDuplicataPattern(img *raster.RasterImage) bool {
	var pattern = raster.NewRasterPattern(512, 280)
	pattern.AddBlackPoints([]image.Point{
		{68, 235}, {70, 239}, {68, 245}, {69, 264}, {74, 242}, {84, 258}, {88, 261}, {90, 255},
		{90, 237}, {89, 260}, {100, 236}, {101, 249}, {100, 264}, {109, 236}, {111, 250}, {109, 264},
		{115, 264}, {190, 249}, {208, 249}, {218, 238}, {236, 241}, {247, 241}, {247, 260}, {265, 243},
		{275, 248}, {299, 254}, {347, 238}, {372, 236}, {390, 249},
	})
	subImage := img.SelectAll()
	match := pattern.IsMatchAt(subImage, 0, 0)
	return match
}

func searchKitchenOrderLines(input *raster.RasterImage) []*raster.RasterSubImage {
	var SubImage = input.Select(image.Rect(0, 280, 30, input.Height))
	var ProductQtyPattern = raster.NewRasterPattern(30, 50)
	ProductQtyPattern.AddWhiteArea(image.Rect(0, 0, 30, 50))
	ProductQtyPattern.DeleteArea(image.Rect(3, 5, 27, 45))
	ProductQtyPattern.SetBlackRatio(0.05, 0.15)
	matches := ProductQtyPattern.SearchAllMatches(SubImage)
	result := make([]*raster.RasterSubImage, 0, len(matches))
	for i, match := range matches {
		if i != len(matches)-1 {
			startY := match.Area.Min.Y
			endY := matches[i+1].Area.Min.Y
			result = append(result, input.SelectRows(startY, endY))
		} else {
			result = append(result, input.SelectRows(match.Area.Min.Y, input.Height-40))
		}
	}
	return result
}

func getCancelImg() *raster.RasterImage {
	const cancelPNG = `iVBORw0KGgoAAAANSUhEUgAAAgAAAABQAQMAAABVp19nAAAABlBMVEX///8AAABVwtN+AAAB7ElE
QVR4nOzWsW7bPBAHcPrToG8qO2YwwqJP4LWAa459jT5CsmkQxAQdNOqN2hQdOHbuZoNAvVLbCT0c
C5uMSjGx6jZAgAIk/otk3s+m72yIuSeuDGQgAxk4Hxgy8PwAiOk1vXkqIM8BqPhVkAAuXGO5XU8j
IwDYg4I/A+yVSgoGZ1CS2KPbBEHYNBFgMAWsYhVHvsPNOgBdnyYCND0AGsYK0Hdwuf7q77cmTQS0
pApcCle4DwH4XhO/hf23H+Mn0CZNBHRUM/y/REYsAIdD7UB3g7sMgNFpYsDBSi0KWNUvld+/Rel6
5F9AMe539l2aCOAOULXGYlVLDwBJPZAwqPbhrSxPEwHiUNCZHiuKAX+SAIBIMwHcKaCWAUCZZgqg
7MzO2b8GiMnOGGfZSWD2CNzhb4HZLnQOmiPQ30/iCDT3wK5NMwEsPQIc2jgCs4PU0vWxCw5GYCP1
gNxg81afMcqartG3cQQ+Sj2ANtisAjD7a/yMVyC7wyQqNQLtAHuDzet3fie3aSJgu6reS76DV3gR
/j/gkygH0D00vDzeIAFJYgBY9d+yuKkXxCq/HyQrB+SWkfBfN5Tbi2lEBNCCXiyLG1W4W/AAqrtD
F7Cg0IV54NGlT78UrX/hASMDGchABjKQgQw8L/AzAAD//3QuxCmigwBwAAAAAElFTkSuQmCC`
	pngData, _ := base64.StdEncoding.DecodeString(cancelPNG)
	pngImg, _ := png.Decode(bytes.NewReader(pngData))
	img := raster.NewRasterImageFromImage(pngImg)
	return img
}
