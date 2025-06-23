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
		var isDabao = isTable31Pattern(input)
		if isKitchenCancelPattern(input) {
			input = input.WithPaste(cancelImg, 0, 160)
			return input.AddMarginBottom(120)
		} else if isKitchenAddPattern(input) {
			header = input.Select(image.Rect(0, 110, input.Width, 190)).Copy()
		} else if isKitchenDuplicataPattern(input) {
			return nil // 取消打印
		} else {
			return input.AddMarginBottom(120)
		}
		var orderLines = searchKitchenOrderLines(input)
		header = header.AddMarginBottom(30)
		header.WithDrawText(time.Now().Format("01/02 15:04"), 0, 50)
		if isDabao {
			// header.WithDrawInvertText("DABAO", 300, 50)
		}
		for _, line := range orderLines {
			input = input.WithAppend(header).WithCutline()
			product := line.Copy().WithScaleY(2)
			input = input.WithAppend(product)
		}
		return input.AddMarginBottom(120)
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
			result = append(result, input.SelectRows(match.Area.Min.Y, input.Height-40))
		}
	}
	return result
}

func isTable31Pattern(img *raster.RasterImage) bool {
	var pattern = raster.NewRasterPattern(512, 240)
	pattern.AddBlackPoints([]image.Point{
		{226, 124}, {232, 125}, {232, 132}, {233, 140}, {253, 128}, {232, 143},
		{257, 124}, {263, 131}, {259, 134}, {264, 140}, {233, 135},
		{172, 136},
	})
	subImage := img.SelectAll()
	match := pattern.IsMatchAt(subImage, 0, 0)
	return match
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

func getTable209() *raster.RasterImage {
	const table209PNG = `iVBORw0KGgoAAAANSUhEUgAAAIAAAAAeAQMAAADenIS7AAAABlBMVEX///8AAABVwtN+AAABFklE
QVR4nKzRMUs7MRgG8Cd980/vDy82g2DAWjs7ZTxQPFfB1VlHoYOO3cxZuE/h0FHq0rkgXBc/QycN
uKnDfYDDyOXgwL0Zf3nyJLzB9pYQgHN/Ye/rcxACIFB6mt++b45f3jjuuXV9UE0m33BlBBIz1qbp
MEhigpRCA4LBooFC9WCERhZ+/sdSlgQDDQvFzwuAjCTY3dV87B4S4QCxZKLyMFRUeo5wJ6nvoTEa
TtvEhgkpNPh0GBP0GEst8IEkXsuKwEerJ8p9e6TgHpL7UI+QttC8lHGD7PIqAlGhYHABtfAmJvKZ
GqfiFT3YFrCuM0+5kx1cL/+d7aDqn1stuilOcdIP3u53M7WQElYPtvNlvwEAAP//cVJEiXmv8XwA
AAAASUVORK5CYII=
`
	pngData, _ := base64.StdEncoding.DecodeString(table209PNG)
	pngImg, _ := png.Decode(bytes.NewReader(pngData))
	img := raster.NewRasterImageFromImage(pngImg)
	return img
}
