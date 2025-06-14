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

func NewRasterImage(width, height int, content []byte) *RasterImage {
	if len(content) != width*height/8 {
		content = make([]byte, width*height/8) // 如果没有内容，初始化为全0的字节切片
	}
	return &RasterImage{
		Width:   width,
		Height:  height,
		Align:   "center", // 默认居中对齐',
		Content: content,
	}
}

// String 返回RasterImage的字符串表示
// 包含宽度、高度、对齐方式和内容长度
// 如果图像为nil，则返回"RasterImage(nil)"
// 注意：内容长度是字节数，不是像素数
// 例如：RasterImage(Width: 384, Height: 240, Align: center, Content Length: 1200)
// 如果图像的宽度或高度为0，则内容长度也会为0
// 如果图像的内容为nil，则内容长度也会为0
// 如果图像的对齐方式为空，则显示为"Align: "（没有值）
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
	// 检查坐标是否在图像范围内
	if x < 0 || x >= img.Width || y < 0 || y >= img.Height {
		return color.White
	}
	// 计算像素所在的字节和位
	byteIndex := (y * img.Width / 8) + (x / 8)
	bitIndex := 7 - (x % 8)
	if img.Content[byteIndex]&(1<<bitIndex) != 0 {
		return color.Black // 黑色像素
	}
	return color.White // 白色像素
}

// GetPixel 返回指定坐标的像素值
// x, y: 像素的坐标
// 返回值：1表示黑色像素，0表示白色像素
// 如果坐标超出图像范围，则返回0（白色）
// 注意：坐标从左上角开始，x为水平坐标，y为垂直坐标
func (img *RasterImage) GetPixel(x, y int) int {
	// 支持负数索引
	if x < 0 {
		x = img.Width + x
	}
	if y < 0 {
		y = img.Height + y
	}
	// 检查坐标是否在图像范围内
	if x < 0 || x >= img.Width || y < 0 || y >= img.Height {
		return 0 // 超出范围，返回白色（0）
	}
	// 计算像素所在的字节和位
	byteIndex := (y * img.Width / 8) + (x / 8)
	bitIndex := 7 - (x % 8)
	if img.Content[byteIndex]&(1<<bitIndex) != 0 {
		return 1 // 黑色像素
	}
	return 0 // 白色像素
}

// SetPixel 设置指定坐标的像素值
// x, y: 像素的坐标
// value: 像素值，1表示黑色，0表示白色
// 如果坐标超出图像范围，则不做任何操作
func (img *RasterImage) setPixel(x, y, value int) {
	// 支持负数索引
	if x < 0 {
		x = img.Width + x
	}
	if y < 0 {
		y = img.Height + y
	}
	// 检查坐标是否在图像范围内
	if x < 0 || x >= img.Width || y < 0 || y >= img.Height {
		return // 超出范围，不做任何操作
	}
	// 计算像素所在的字节和位
	byteIndex := (y * img.Width / 8) + (x / 8)
	bitIndex := 7 - (x % 8)
	if value == 1 {
		img.Content[byteIndex] |= (1 << bitIndex) // 设置为黑色
	} else {
		img.Content[byteIndex] &^= (1 << bitIndex) // 设置为白色
	}
}

// SetPixel 设置指定坐标的像素为白色
func (img *RasterImage) SetPixelWhite(x, y int) {
	img.setPixel(x, y, 0) // 设置为白色
}

// SetPixel 设置指定坐标的像素为黑色
func (img *RasterImage) SetPixelBlack(x, y int) {
	img.setPixel(x, y, 1) // 设置为黑色
}

// IsAllBlack 检查图像是否全黑
// 如果图像为nil或内容为nil，则返回false
// 如果图像的内容全部为0xFF（全黑），则返回true
// 否则返回false
// 注意：全黑的定义是所有字节都为0xFF，即每个像素都是黑色
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

// IsAllWhite 检查图像是否全白
// 如果图像为nil或内容为nil，则返回false
// 如果图像的内容全部为0x00（全白），则返回true
// 否则返回false
// 注意：全白的定义是所有字节都为0x00，即每个像素都是白色
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

// IsWhiteLine 检查指定行是否全是白色
// y: 行索引，从0开始
func (img *RasterImage) IsWhiteLine(y int) bool {
	if img == nil || img.Content == nil || y < 0 || y >= img.Height {
		return false // 无效图像或行索引
	}
	for x := 0; x < img.Width; x++ {
		if img.GetPixel(x, y) != 0 { // 如果有任何像素不是白色，则返回false
			return false
		}
	}
	return true // 所有像素都是白色的
}

// IsWhiteColumn 检查指定列是否全是白色
// x: 列索引，从0开始
func (img *RasterImage) IsWhiteColumn(x int) bool {
	if img == nil || img.Content == nil || x < 0 || x >= img.Width {
		return false // 无效图像或列索引
	}
	for y := 0; y < img.Height; y++ {
		if img.GetPixel(x, y) != 0 { // 如果有任何像素不是白色，则返回false
			return false
		}
	}
	return true // 所有像素都是白色的
}
