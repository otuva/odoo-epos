package raster

import (
	"image"
	"iter"
	"sort"
)

type RasterOCR struct {
	PngBase64     string
	CharAreas     map[string]image.Rectangle
	templateImage *RasterImage
}

func (ocr *RasterOCR) CharImages() iter.Seq2[string, *RasterSubImage] {
	if ocr.templateImage == nil {
		ocr.templateImage = NewRasterImageFromBase64(ocr.PngBase64)
	}
	// 稳定排序
	keys := make([]string, 0, len(ocr.CharAreas))
	for k := range ocr.CharAreas {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return func(yield func(string, *RasterSubImage) bool) {
		for _, s := range keys {
			area := ocr.CharAreas[s]
			img := ocr.templateImage.Select(area)
			if img == nil {
				continue
			}
			if !yield(s, img) {
				return
			}
		}

	}
}

func (ocr *RasterOCR) Recognize(img *RasterSubImage) string {
	return ""
}
