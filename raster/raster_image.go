package raster

import (
	"fmt"
	"image"
	"image/color"
)

// EPOSImage 表示图片数据
type RasterImage struct {
	Width    int    `xml:"width,attr"`
	Height   int    `xml:"height,attr"`
	Align    string `xml:"align,attr"`
	Content  []byte `xml:",chardata"` // 图片数据
	filename string // 可选的文件名，用于保存图片时使用
}

func NewRasterImage(width, height int) *RasterImage {
	width = (width + 7) &^ 7 // 向上取整，确保宽度是8的倍数
	rowBytes := width / 8    // 只适用于宽度为8的倍数
	content := make([]byte, height*rowBytes)

	return &RasterImage{
		Width:   width,
		Height:  height,
		Align:   "center",
		Content: content,
	}
}

func (img *RasterImage) String() string {
	if img == nil {
		return "RasterImage(nil)"
	}
	return fmt.Sprintf("RasterImage(Width: %d, Height: %d, Align: %s, Content Length: %d)",
		img.Width, img.Height, img.Align, len(img.Content))
}

func (img *RasterImage) ColorModel() color.Model {
	// 返回黑白二值图像的颜色模型
	return color.Palette{color.White, color.Black}
}

func (img *RasterImage) Bounds() image.Rectangle {
	// 返回图像的边界矩形
	return image.Rect(0, 0, img.Width, img.Height)
}

func (img *RasterImage) At(x, y int) color.Color {
	if img.GetPixel(x, y) == 1 {
		return color.Black // 黑色像素
	}
	return color.White // 白色像素
}

// GetPixel 返回指定坐标的像素值
// x, y: 像素的坐标
// 返回值：1表示黑色像素，0表示白色像素
// 如果坐标超出图像范围，则返回0（白色）
func (img *RasterImage) GetPixel(x, y int) int {
	if img.Width%8 != 0 {
		panic("RasterImage width must be a multiple of 8 !")
	}
	if x < 0 {
		x = img.Width + x
	}
	if y < 0 {
		y = img.Height + y
	}
	if x < 0 || x >= img.Width || y < 0 || y >= img.Height {
		return 0 // 超出范围，返回白色
	}
	rowBytes := img.Width / 8
	byteIndex := y*rowBytes + (x / 8)
	bitIndex := 7 - (x % 8)
	if img.Content[byteIndex]&(1<<bitIndex) != 0 {
		return 1 // 黑色像素
	}
	return 0 // 白色像素
}

func (img *RasterImage) GetRow(y int) []byte {
	if img.Width%8 != 0 {
		panic("RasterImage width must be a multiple of 8 !")
	}
	if y < 0 || y >= img.Height {
		return nil
	}
	rowBytes := img.Width / 8
	byteIndex := y * rowBytes
	if byteIndex+rowBytes > len(img.Content) {
		return nil
	}
	return img.Content[byteIndex : byteIndex+rowBytes]
}

// SetPixel 设置指定坐标的像素值（只考虑宽度为8的倍数）
func (img *RasterImage) SetPixel(point image.Point, value int) {
	if img.Width%8 != 0 {
		panic("RasterImage width must be a multiple of 8 !")
	}
	x := point.X
	y := point.Y
	if x < 0 {
		x = img.Width + x
	}
	if y < 0 {
		y = img.Height + y
	}
	if x < 0 || x >= img.Width || y < 0 || y >= img.Height {
		return
	}
	rowBytes := img.Width / 8
	byteIndex := y*rowBytes + (x / 8)
	bitIndex := 7 - (x % 8)
	if value == 1 {
		img.Content[byteIndex] |= (1 << bitIndex)
	} else {
		img.Content[byteIndex] &^= (1 << bitIndex)
	}
}

// SetPixel 设置指定坐标的像素为白色
func (img *RasterImage) SetPixelWhite(x, y int) {
	img.SetPixel(image.Point{x, y}, 0) // 设置为白色
}

// SetPixel 设置指定坐标的像素为黑色
func (img *RasterImage) SetPixelBlack(x, y int) {
	img.SetPixel(image.Point{x, y}, 1) // 设置为黑色
}

func (img *RasterImage) IsAllBlack() bool {
	if img == nil || img.Content == nil {
		return false
	}
	for _, b := range img.Content {
		if b != 0xFF { // 如果有任何字节不是全黑，则返回false
			return false
		}
	}
	return true // 所有字节都是全黑
}

func (img *RasterImage) IsAllWhite() bool {
	if img == nil || img.Content == nil {
		return false
	}
	for _, b := range img.Content {
		if b != 0x00 { // 如果有任何字节不是全白，则返回false
			return false
		}
	}
	return true // 所有字节都是全白
}
