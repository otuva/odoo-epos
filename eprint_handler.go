package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
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
	var img image.Image
	if strings.HasPrefix(url, "file:") {
		// 处理本地文件路径
		filePath := strings.TrimPrefix(url, "file:")
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open local file: %w", err)
		}
		defer file.Close()
		img, err = png.Decode(file)
		if err != nil {
			return nil, fmt.Errorf("failed to decode local PNG file: %w", err)
		}
		return img, nil
	} else if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to download png: status %d", resp.StatusCode)
		}
		img, err = png.Decode(resp.Body)
		if err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(url, "data:image/png;base64,") {
		// 处理Base64编码的PNG数据
		base64Data := strings.TrimPrefix(url, "data:image/png;base64,")
		data, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 PNG data: %w", err)
		}
		img, err = png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to decode PNG data: %w", err)
		}
	} else {
		return nil, fmt.Errorf("unsupported URL scheme: %s", url)
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

func ePrintLocalPNGhandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"success":false,"msg":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var printerName, filePath string
	printerName = r.URL.Query().Get("printer")
	filePath = r.URL.Query().Get("path")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// 检查printerName和filePath参数是否存在
	if printerName == "" || filePath == "" {
		http.Error(w, `{"success":false,"msg":"Missing printer or dir parameter"}`, http.StatusBadRequest)
		return
	}

	// 读取filePath文件夹下的所有png文件
	files, err := ReadPNGFilesFromPath(filePath)
	if err != nil || len(files) == 0 {
		http.Error(w, `{"success":false,"msg":"Failed to read PNG files: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	printer, ok := Printers[printerName]
	if !ok {
		http.Error(w, `{"success":false,"msg":"Printer not found"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	// 逐个打印文件
	fmt.Fprintf(w, `<h1><pre>Starting to print %d files from %s</pre></h1>`, len(files), filePath)
	flusher, _ := w.(http.Flusher)
	for _, file := range files {
		select {
		case <-r.Context().Done():
			// 客户端已断开连接，提前结束
			return
		default:
			fmt.Fprintf(w, `<pre>{"success":true,"msg":"Printing file %s"}</pre>`, file)
			if flusher != nil {
				flusher.Flush() // 确保及时发送响应到客户端
			}
		}
		time.Sleep(2 * time.Second) // 模拟打印延时，实际打印时可以去掉
		pngFile, err := os.Open(file)
		if err != nil {
			fmt.Fprintf(w, `<pre style="color:red;">{"success":false,"msg":"Failed to open file %s: %s"}</pre>`, file, err.Error())
			continue // 如果打开某个文件失败，跳过该文件
		}
		defer pngFile.Close()
		pngImg, err := png.Decode(pngFile)
		if err != nil {
			fmt.Fprintf(w, `<pre style="color:red;">{"success":false,"msg":"Failed to decode PNG file %s: %s"</pre>}`, file, err.Error())
			continue // 如果解码某个文件失败，跳过该文件
		}
		img := NewRasterImageFromPNG(pngImg)
		err = printer.PrintRasterImage(img)
		if err != nil {
			fmt.Fprintf(w, `<pre style="color:red;">{"success":false,"msg":"Failed to print file %s: %s"}</pre>`, file, err.Error())
			continue // 如果打印某个文件失败，跳过该文件
		}
		fmt.Fprintf(w, `<pre style="color:green;">{"success":true,"msg":"File %s printed successfully "}</pre>`, file)
		// 刷新响应，确保每个文件打印后客户端能及时接收到响应
		if flusher != nil {
			flusher.Flush()
		}
	}
	w.Write([]byte(`<h1 style="color:blue;"><pre>Print Complete</pre></h1>`))
}

// ReadPNGFilesFromPath 读取指定路径下的所有PNG文件，并返回文件路径列表
func ReadPNGFilesFromPath(path string) ([]string, error) {
	var pngFiles []string
	err := filepath.Walk(path, func(fp string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".png") {
			pngFiles = append(pngFiles, fp)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// 按文件名排序
	sort.Strings(pngFiles)
	return pngFiles, nil
}
