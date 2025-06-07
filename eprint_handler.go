package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
)

func ePrintPNGhandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var printerName, pngUrl string
	switch r.Method {
	case http.MethodGet:
		printerName = r.URL.Query().Get("x_printer")
		pngUrl = r.URL.Query().Get("x_url")
	case http.MethodPost:
		var data map[string]string
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, `{"success":false,"msg":"Invalid request body"}`, http.StatusBadRequest)
			return
		}
		printerName = data["x_printer"]
		pngUrl = data["x_url"]
	default:
		http.Error(w, `{"success":false,"msg":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if printerName == "" || pngUrl == "" {
		http.Error(w, `{"success":false,"msg":"Missing printer or url parameter"}`, http.StatusBadRequest)
		return
	}

	printer, ok := Printers[printerName]
	if !ok {
		http.Error(w, `{"success":false,"msg":"Printer not found"}`, http.StatusBadRequest)
		return
	}

	pngImg, err := DownloadPNGImage(pngUrl)
	if err != nil {
		http.Error(w, `{"success":false,"msg":"Failed to download PNG image: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	img := NewRasterImageFromPNG(pngImg)
	if img == nil {
		http.Error(w, `{"success":false,"msg":"Failed to create raster image from PNG"}`, http.StatusInternalServerError)
		return
	}
	if err := printer.PrintRasterImage(img); err != nil {
		http.Error(w, `{"success":false,"msg":"Failed to print raster image: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true,"msg":"Image printed successfully"}`))
}

func ePrintRAWhandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var printerName, rawHex string
	switch r.Method {
	case http.MethodGet:
		printerName = r.URL.Query().Get("x_printer")
		rawHex = r.URL.Query().Get("x_hex")
	case http.MethodPost:
		var data map[string]string
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, `{"success":false,"msg":"Invalid request body"}`, http.StatusBadRequest)
			return
		}
		printerName = data["x_printer"]
		rawHex = data["x_hex"]
	default:
		http.Error(w, `{"success":false,"msg":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if printerName == "" || rawHex == "" {
		http.Error(w, `{"success":false,"msg":"Missing printer or data parameter"}`, http.StatusBadRequest)
		return
	}

	printer, ok := Printers[printerName]
	if !ok {
		http.Error(w, `{"success":false,"msg":"Printer not found"}`, http.StatusBadRequest)
		return
	}

	rawBytes, err := hex.DecodeString(rawHex)
	if err != nil {
		http.Error(w, `{"success":false,"msg":"Invalid hex data: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	if err := printer.PrintRaw(rawBytes); err != nil {
		http.Error(w, `{"success":false,"msg":"Failed to print raw commands: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true,"msg":"Raw commands printed successfully"}`))
}

// DownloadPngImage 从指定URL下载PNG图片并解码为image.Image
func DownloadPNGImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download png: status %d", resp.StatusCode)
	}
	img, err := png.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// DitherImage 对图像进行Floyd-Steinberg抖动处理，输出黑白图像
func DitherImage(src image.Image) *image.Gray {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	grayImg := image.NewGray(bounds)

	// 先将原图像素拷贝到灰度图
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, _ := src.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			gray := uint8((r*299 + g*587 + b*114 + 500) / 1000 >> 8) // 0~255
			grayImg.SetGray(bounds.Min.X+x, bounds.Min.Y+y, color.Gray{Y: gray})
		}
	}

	// Floyd-Steinberg 抖动
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			oldPixel := grayImg.GrayAt(bounds.Min.X+x, bounds.Min.Y+y).Y
			newPixel := uint8(255)
			if oldPixel < 230 { // 阈值可根据打印效果调整
				newPixel = 0
			}
			grayImg.SetGray(bounds.Min.X+x, bounds.Min.Y+y, color.Gray{Y: newPixel})
			quantError := int16(oldPixel) - int16(newPixel)

			// 误差扩散
			if x+1 < w {
				grayImg.SetGray(bounds.Min.X+x+1, bounds.Min.Y+y, addError(grayImg.GrayAt(bounds.Min.X+x+1, bounds.Min.Y+y).Y, quantError*7/16))
			}
			if x-1 >= 0 && y+1 < h {
				grayImg.SetGray(bounds.Min.X+x-1, bounds.Min.Y+y+1, addError(grayImg.GrayAt(bounds.Min.X+x-1, bounds.Min.Y+y+1).Y, quantError*3/16))
			}
			if y+1 < h {
				grayImg.SetGray(bounds.Min.X+x, bounds.Min.Y+y+1, addError(grayImg.GrayAt(bounds.Min.X+x, bounds.Min.Y+y+1).Y, quantError*5/16))
			}
			if x+1 < w && y+1 < h {
				grayImg.SetGray(bounds.Min.X+x+1, bounds.Min.Y+y+1, addError(grayImg.GrayAt(bounds.Min.X+x+1, bounds.Min.Y+y+1).Y, quantError*1/16))
			}
		}
	}
	return grayImg
}

// addError 辅助函数，像素值加上误差并裁剪到0~255
func addError(val uint8, err int16) color.Gray {
	v := int16(val) + err
	if v < 0 {
		v = 0
	}
	if v > 255 {
		v = 255
	}
	return color.Gray{Y: uint8(v)}
}
