package main

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
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

func NewRasterImageFromPNG(img image.Image) *RasterImage {
	if img == nil {
		return nil
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	// 保证宽度为8的倍数
	if width%8 != 0 {
		width = (width/8 + 1) * 8
	}
	content := make([]byte, height*width/8)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			byteIndex := y*(width/8) + x/8
			bitIndex := 7 - (x % 8)
			var isBlack bool
			if x < bounds.Dx() {
				r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
				// 计算灰度值，范围0-65535
				gray := (r*299 + g*587 + b*114) / 1000
				isBlack = gray < 48000 // 只有灰度非常高（接近65535）才算白，其余都算黑
			} else {
				isBlack = false // 超出原图部分补白
			}
			if isBlack {
				content[byteIndex] |= 1 << bitIndex
			}
		}
	}

	return &RasterImage{
		Width:   width,
		Height:  height,
		Align:   "center", // 默认居中对齐
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

// WithCrop 返回裁剪后的图像
// x, y: 裁剪区域左上角坐标
// width, height: 裁剪区域的宽度和高度
// 如果裁剪区域超出原图范围，则返回nil
// 如果裁剪区域无效（如宽度或高度为0），也返回nil
// 注意：裁剪后的图像宽度和高度必须是8的倍数
// 返回的新图像对象的宽度和高度可能会被调整为8的倍数
// 如果原图像无效（如宽度或高度小于等于0，或内容为nil），也返回nil
// 如果裁剪区域的宽度或高度不是8的倍数，则会自动调整为最接近的8的倍数
// 如果裁剪区域的左上角坐标超出原图范围，则返回nil
// 如果裁剪区域的宽度或高度为0，则返回nil
// 如果裁剪区域的左上角坐标为负数，则返回nil
// 如果裁剪区域的宽度或高度超过原图范围，则返回nil
// 如果裁剪区域的左上角坐标在原图范围内，但宽度或高度超出原图范围，则返回nil
func (img *RasterImage) WithCrop(x, y, width, height int) *RasterImage {
	if img == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil {
		return nil
	}

	// 支持负数索引
	if x < 0 {
		x = img.Width + x
	}
	if y < 0 {
		y = img.Height + y
	}

	// 检查参数有效性
	if x < 0 || y < 0 || width <= 0 || height <= 0 || x+width > img.Width || y+height > img.Height {
		return nil // 无效的裁剪参数
	}

	croppedContent := make([]byte, height*width/8)
	for row := 0; row < height; row++ {
		srcRowStart := (y+row)*img.Width/8 + x/8
		srcRowEnd := srcRowStart + width/8
		dstRowStart := row * width / 8
		copy(croppedContent[dstRowStart:], img.Content[srcRowStart:srcRowEnd])
	}

	return &RasterImage{
		Width:   width,
		Height:  height,
		Align:   img.Align,
		Content: croppedContent,
	}
}

// WithCropRows 返回裁剪后的图像，仅裁剪行
// startRow, endRow: 裁剪区域的起始行和结束行（包含）
// 如果裁剪区域超出原图范围，则返回nil
// 如果裁剪区域无效（如起始行大于结束行，或行数为0），也返回nil
// 注意：裁剪后的图像宽度和高度必须是8的倍数
// 返回的新图像对象的宽度和高度可能会被调整为8的倍数
// 如果原图像无效（如宽度或高度小于等于0，或内容为nil），也返回nil
// 如果裁剪区域的行数不是8的倍数，则会自动调整为最接近的8的倍数
// 如果裁剪区域的起始行坐标超出原图范围，则返回nil
// 如果裁剪区域的行数为0，则返回nil
func (img *RasterImage) WithCropRows(startRow, endRow int) *RasterImage {
	if img == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil {
		return nil
	}
	// 支持负数索引
	if startRow < 0 {
		startRow = img.Height + startRow
	}
	if endRow < 0 {
		endRow = img.Height + endRow
	}
	// 检查参数有效性
	if startRow < 0 || endRow < startRow || endRow >= img.Height {
		return nil // 无效的行范围
	}

	croppedHeight := endRow - startRow + 1
	croppedContent := make([]byte, croppedHeight*img.Width/8)
	for row := 0; row < croppedHeight; row++ {
		srcRowStart := (startRow + row) * img.Width / 8
		srcRowEnd := srcRowStart + img.Width/8
		dstRowStart := row * img.Width / 8
		copy(croppedContent[dstRowStart:], img.Content[srcRowStart:srcRowEnd])
	}

	return &RasterImage{
		Width:   img.Width,
		Height:  croppedHeight,
		Align:   img.Align,
		Content: croppedContent,
	}
}

// WithAppend 返回拼接后的图像
// other: 要拼接的另一个图像
// 如果原图像或其他图像无效（如宽度或高度小于等于0，或内容为nil），则返回nil
// 如果其他图像的宽度与原图像不匹配，则返回nil
// 拼接后的图像高度为原图像高度加上其他图像高度
// 拼接后的图像宽度与原图像相同
func (img *RasterImage) WithAppend(other *RasterImage) *RasterImage {
	if img == nil || other == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil {
		return nil
	}
	if other.Width != img.Width {
		return nil // 宽度不匹配，无法拼接
	}

	newHeight := img.Height + other.Height
	newContent := make([]byte, newHeight*img.Width/8)
	copy(newContent, img.Content)
	copy(newContent[img.Height*img.Width/8:], other.Content)

	return &RasterImage{
		Width:   img.Width,
		Height:  newHeight,
		Align:   img.Align,
		Content: newContent,
	}
}

// WithPaste 返回粘贴后的图像
// other: 要粘贴的另一个图像
// x, y: 粘贴位置的左上角坐标
// 如果粘贴位置超出原图像范围，则返回原图像
// 如果其他图像的宽度或高度超过原图像范围，则自动裁剪
// 粘贴后的图像宽度和高度与原图像相同
// 粘贴操作会将其他图像的内容覆盖到原图像的指定位置
// 注意：粘贴操作不会改变原图像的对齐方式
// 如果粘贴位置在原图像范围内，但其他图像的内容超出原图像范围，则只粘贴在原图像范围内的部分
// 如果其他图像的内容为空，则不会对原图像进行任何修改
// 如果其他图像的宽度或高度为0，则不会对原图像进行任何修改
// 如果其他图像的内容为nil，则不会对原图像进行任何修改
// 如果其他图像的宽度或高度小于0，则不会对原图像进行任何修改
func (img *RasterImage) WithPaste(other *RasterImage, x, y int) *RasterImage {
	// 检查参数有效性
	if img == nil || other == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil || other.Content == nil {
		return img
	}
	// 支持负数索引
	if x < 0 {
		x = img.Width + x
	}
	if y < 0 {
		y = img.Height + y
	}

	if x < 0 || y < 0 || x >= img.Width || y >= img.Height {
		return img // 粘贴起点超出原图范围
	}

	// 自动裁剪other的宽高，防止越界
	maxWidth := img.Width - x
	maxHeight := img.Height - y
	pasteWidth := other.Width
	pasteHeight := other.Height
	if pasteWidth > maxWidth {
		pasteWidth = maxWidth
	}
	if pasteHeight > maxHeight {
		pasteHeight = maxHeight
	}

	newContent := make([]byte, len(img.Content))
	copy(newContent, img.Content)

	for row := 0; row < pasteHeight; row++ {
		srcRowStart := row * other.Width / 8
		dstRowStart := (y+row)*img.Width/8 + x/8
		for col := 0; col < pasteWidth; col++ {
			if (other.Content[srcRowStart+col/8] & (1 << (7 - (col % 8)))) != 0 {
				newContent[dstRowStart+col/8] |= 1 << (7 - ((x + col) % 8))
			}
		}
	}

	return &RasterImage{
		Width:   img.Width,
		Height:  img.Height,
		Align:   img.Align,
		Content: newContent,
	}
}

// WithErase 返回擦除后的图像
// x, y: 擦除区域的左上角坐标
// width, height: 擦除区域的宽度和高度
// 如果擦除区域超出原图像范围，则返回原图像
// 如果擦除区域无效（如宽度或高度为0），也返回原图像
// 擦除操作会将指定区域的内容清零
// 注意：擦除操作不会改变原图像的对齐方式
// 如果擦除区域在原图像范围内，但宽度或高度超出原图像范围，则只擦除在原图像范围内的部分
// 如果擦除区域的左上角坐标超出原图像范围，则返回原图像
// 如果擦除区域的宽度或高度为0，则返回原图像
// 如果擦除区域的左上角坐标为负数，则返回原图像
func (img *RasterImage) WithErase(x, y, width, height int) *RasterImage {
	// 检查参数有效性
	if img == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil {
		return img
	}
	// 支持负数索引
	if x < 0 {
		x = img.Width + x
	}
	if y < 0 {
		y = img.Height + y
	}

	if x < 0 || y < 0 || width <= 0 || height <= 0 || x+width > img.Width || y+height > img.Height {
		return img // 无效的擦除参数
	}

	newContent := make([]byte, len(img.Content))
	copy(newContent, img.Content)

	for row := 0; row < height; row++ {
		srcRowStart := (y + row) * img.Width / 8
		dstRowStart := srcRowStart + x/8
		for col := 0; col < width; col++ {
			newContent[dstRowStart+col/8] &^= (1 << (7 - ((x + col) % 8))) // 直接清零，无需判断
		}
	}

	return &RasterImage{
		Width:   img.Width,
		Height:  img.Height,
		Align:   img.Align,
		Content: newContent,
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
