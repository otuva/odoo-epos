package raster

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// 从png文件创建RasterImage对象
func NewRasterImageFromFile(filePath string) *RasterImage {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Failed to open file:", err)
		return nil
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		fmt.Println("Failed to decode PNG image:", err)
		return nil
	}

	result := NewRasterImageFromImage(img)
	result.filename = filePath // 保存文件名以便后续使用

	return result
}

func NewRasterImageFromBase64(base64Str string) *RasterImage {
	pngData, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		fmt.Println("Failed to decode base64 string:", err)
		return nil
	}
	pngImg, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		fmt.Println("Failed to decode PNG image from base64:", err)
		return nil
	}
	// 创建RasterImage对象
	img := NewRasterImageFromImage(pngImg)
	return img
}

func (img *RasterImage) GetFilename() string {
	if img == nil {
		return ""
	}
	return img.filename
}

func NewRasterImageFromImage(img image.Image) *RasterImage {
	if img == nil {
		return nil
	}
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()
	width := (imgWidth + 7) &^ 7 // 向上取整，确保宽度是8的倍数
	height := imgHeight          // 高度保持不变
	content := make([]byte, height*width/8)
	rs := &RasterImage{
		Width:   width,
		Height:  height,
		Align:   "center", // 默认居中对齐
		Content: content,
	}

	for y := range imgHeight {
		for x := range imgWidth {
			var isBlack bool
			r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			// 计算灰度值，范围0-65535
			gray := (r*299 + g*587 + b*114) / 1000
			isBlack = gray < 48000 // 只有灰度非常高（接近65535）才算白，其余都算黑

			if isBlack {
				rs.SetPixelBlack(x, y)
			}
		}
	}
	return rs
}

// 将图像转换为1bit黑白image.Image接口
func (img *RasterImage) ToPngImage() image.Image {
	if img == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil {
		return nil
	}
	bounds := image.Rect(0, 0, img.Width, img.Height)
	palette := []color.Color{color.White, color.Black}
	palettedImg := image.NewPaletted(bounds, palette)
	widthBytes := img.Width / 8

	for y := 0; y < img.Height; y++ {
		srcRow := img.Content[y*widthBytes : (y+1)*widthBytes]
		dstRow := palettedImg.Pix[y*palettedImg.Stride : y*palettedImg.Stride+img.Width]
		for x := 0; x < img.Width; x++ {
			b := srcRow[x/8]
			if (b & (1 << (7 - (x % 8)))) != 0 {
				dstRow[x] = 1 // 黑色
			} else {
				dstRow[x] = 0 // 白色
			}
		}
	}
	return palettedImg
}

// 将二值矩形图像输出到png文件采用无压缩
func (img *RasterImage) SaveToPngFile(filePath string) error {
	// 建输文件
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将图像编码为PNG并写入文件
	return png.Encode(file, img.ToPngImage())
}
