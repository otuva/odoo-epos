package main

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"image"
	"image/png"
	"os"
	"strconv"
	"strings"
)

// EPOSImage 表示图片数据
type RasterImage struct {
	Width   int    `xml:"width,attr"`
	Height  int    `xml:"height,attr"`
	Align   string `xml:"align,attr"`
	Content []byte `xml:",chardata"` // 图片数据
}

// NewRasterImage 从XML数据中解析并返回RasterImage对象
func NewRasterImageFromXML(payload []byte) (*RasterImage, error) {
	envelope := &Envelope{}
	if err := xml.Unmarshal(payload, envelope); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SOAP envelope: %w", err)
	}
	return envelope.Body.EposPrint.Image, nil
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

func LowHighValue(value int) (low, high byte) {
	// 计算低位和高位字节
	low = byte(value & 0xFF)
	high = byte((value >> 8) & 0xFF)
	return
}

func (img *RasterImage) toEscPosRasterCommand() []byte {
	// 参数验证
	if img == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil {
		return nil
	}
	const GS = 0x1D // ESC/POS 命令前缀
	// 计算 xL 和 xH
	xL, xH := LowHighValue(img.Width / 8)
	// 计算 yL 和 yH
	yL, yH := LowHighValue(img.Height)

	result := append([]byte{GS, 'v', 48, 0, xL, xH, yL, yH}, img.Content...)

	return result
}

func (img *RasterImage) ToEscPosRasterCommand(maxHeight int) []byte {
	if maxHeight <= 0 || maxHeight > img.Height {
		return img.toEscPosRasterCommand()
	}
	result := make([]byte, 0, 100+len(img.Content))
	remainingHeight := img.Height
	offset := 0
	widthBytes := img.Width / 8
	const GS = 0x1D                       // ESC/POS 命令前缀
	xL, xH := LowHighValue(img.Width / 8) // 宽度不变，循环外计算
	for remainingHeight > 0 {
		currentHeight := min(remainingHeight, maxHeight)
		startPos := offset * widthBytes
		endPos := min(startPos+(currentHeight*widthBytes), len(img.Content))
		currentContent := img.Content[startPos:endPos]
		yL, yH := LowHighValue(currentHeight)
		command := append([]byte{GS, 'v', 48, 0, xL, xH, yL, yH}, currentContent...)
		result = append(result, command...)
		remainingHeight -= currentHeight
		offset += currentHeight
	}
	return result
}

func (img *RasterImage) AddMarginLeft(margin int) {
	if margin <= 0 {
		return // 如果边距小于等于0，则不添加边距
	}
	widthBytes := img.Width / 8
	marginBytes := margin / 8
	newWidth := img.Width + margin
	newContent := make([]byte, img.Height*(newWidth/8))
	for row := 0; row < img.Height; row++ {
		oldRowStart := row * widthBytes
		oldRowEnd := oldRowStart + widthBytes
		newRowStart := row * (newWidth / 8)
		// 前 marginBytes 字节自动为0（空白），直接拷贝原内容到新行后面
		copy(newContent[newRowStart+marginBytes:], img.Content[oldRowStart:oldRowEnd])
	}
	img.Width = newWidth
	img.Content = newContent
}

func (img *RasterImage) AddMarginBottom(margin int) {
	if margin < 0 {
		margin = 0
	}
	img.Height += margin
	newContent := make([]byte, len(img.Content)+margin*img.Width/8)
	copy(newContent, img.Content)
	img.Content = newContent
}

func (img *RasterImage) AutoLeftMargin(width int) int {
	if img.Width >= width {
		return 0 // 如果图像宽度大于等于指定宽度，则不需要左边距
	}
	align := strings.ToLower(img.Align)
	margin := 0 // 默认左边距为0
	switch align {
	case "left":
		return 0 // 左对齐不需要左边距
	case "right":
		margin = width - img.Width // 右对齐需要填充到指定宽度
	default:
		// 如果对齐方式不明确，默认使用居中对齐
		margin = (width - img.Width) / 2
	}
	margin = (margin / 8) * 8 // 确保左边距是8的倍数
	return margin
}

// AddMargin 添加左边距和下边距
func (img *RasterImage) AddMargin(marginLeft, marginBottom int) {
	if marginLeft < 0 {
		marginLeft = 0
	}
	if marginBottom < 0 {
		marginBottom = 0
	}
	img.AddMarginLeft(marginLeft)
	img.AddMarginBottom(marginBottom)
}

// 将图像转换为image.Image接口
func (img *RasterImage) ToPngImage() image.Image {
	if img == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil {
		return nil
	}
	bounds := image.Rect(0, 0, img.Width, img.Height)
	binaryImg := image.NewGray(bounds)
	widthBytes := img.Width / 8

	for y := range img.Height {
		for x := range img.Width {
			byteIndex := y*widthBytes + x/8
			bitIndex := 7 - (x % 8)
			grayIdx := y*img.Width + x
			if byteIndex < len(img.Content) && (img.Content[byteIndex]&(1<<bitIndex)) == 0 {
				binaryImg.Pix[grayIdx] = 255 // 白色
			} else {
				binaryImg.Pix[grayIdx] = 0 // 黑色
			}
		}
	}
	return binaryImg
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

// Envelope 表示SOAP信封结构
type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    Body     `xml:"Body"`
}

// Body 表示SOAP消息体
type Body struct {
	XMLName   xml.Name  `xml:"Body"`
	EposPrint EposPrint `xml:"epos-print"`
}

// EposPrint 表示打印机指令容器
type EposPrint struct {
	XMLName xml.Name     `xml:"epos-print"`
	Xmlns   string       `xml:"xmlns,attr"`
	Image   *RasterImage `xml:"image"`
}

func (img *RasterImage) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias RasterImage
	aux := &struct {
		Width   string `xml:"width,attr"`
		Height  string `xml:"height,attr"`
		Content string `xml:",chardata"`
		*Alias
	}{
		Alias: (*Alias)(img),
	}

	if err := d.DecodeElement(&aux, &start); err != nil {
		return err
	}

	var err error
	img.Width, err = strconv.Atoi(aux.Width)
	if err != nil {
		return err
	}

	img.Height, err = strconv.Atoi(aux.Height)
	if err != nil {
		return err
	}

	img.Content, err = base64.StdEncoding.DecodeString(aux.Content)
	if err != nil {
		return err
	}

	return nil
}
