package main

import (
	"fmt"
	"os"
	"time"
)

type FilePrinter string

func (p FilePrinter) String() string {
	return fmt.Sprintf("FilePrinter{Dir: %s}", string(p))
}

func NewFilePrinter(path string) FilePrinter {
	if path == "" {
		path = "./png-receipts" // 默认路径
	} else if path[len(path)-1] == '/' {
		path = path[:len(path)-1] // 去掉末尾的斜杠
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", path, err)
		return FilePrinter("")
	}
	return FilePrinter(path)
}

func (p FilePrinter) OpenCashBox() error {
	return nil // 文件打印机不支持打开钱箱
}
func (p FilePrinter) PrintRasterImage(img *RasterImage) error {
	filename := fmt.Sprintf("%s/%s.png", string(p), time.Now().Format("20060102-150405"))
	return img.SaveToPngFile(filename)
}
