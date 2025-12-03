package main

import (
	"bytes"
	"embed"
	"fmt"
	"image/png"

	eprinter "github.com/xiaohao0576/odoo-epos/printer"
	"github.com/xiaohao0576/odoo-epos/raster"
)

//go:embed png/odoo_logo.png
var ImageFS embed.FS

func PrintTestPage(printer eprinter.EPrinter) error {
	// 从embed文件系统中读取图片数据
	imageData, err := ImageFS.ReadFile("png/odoo_logo.png")
	if err != nil {
		fmt.Println("Failed to read embedded image:", err)
		return err
	}

	// 解码PNG图片数据
	imgReader := bytes.NewReader(imageData)
	pngImg, err := png.Decode(imgReader)
	if err != nil {
		fmt.Println("Failed to decode PNG image:", err)
		return err
	}

	// 使用解码后的图片创建RasterImage
	img := raster.NewRasterImageFromImage(pngImg)

	// Test Image Print
	err = printer.PrintRasterImage(img)
	if err != nil {
		fmt.Println("Failed to print image:", err)
		return err
	}
	fmt.Println("Image print success.", printer)

	// Test Open Cash Box
	err = printer.OpenCashBox()
	if err != nil {
		fmt.Println("Failed to open cash drawer:", err)
		return err
	}
	// Cash drawer opened successfully
	fmt.Println("Cash drawer opened successfully.", printer)

	return nil
}
