package raster

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/bits"
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

// WithDeleteRows 返回删除指定行后的图像
// startRow, endRow: 要删除的行范围（包含）
// 如果删除范围超出原图范围，则返回原图像
// 如果删除范围无效（如起始行大于结束行，或行数为0），也返回原图像
func (img *RasterImage) WithDeleteRows(startRow, endRow int) *RasterImage {
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
	if startRow < 0 {
		startRow = 0
	}
	if endRow >= img.Height {
		endRow = img.Height - 1
	}
	if startRow > endRow || startRow >= img.Height || endRow < 0 {
		return img // 无需删除
	}
	widthBytes := img.Width / 8
	newHeight := img.Height - (endRow - startRow + 1)
	if newHeight <= 0 {
		return nil // 全部删除
	}
	newContent := make([]byte, newHeight*widthBytes)
	dstRow := 0
	for row := 0; row < img.Height; row++ {
		if row < startRow || row > endRow {
			copy(
				newContent[dstRow*widthBytes:(dstRow+1)*widthBytes],
				img.Content[row*widthBytes:(row+1)*widthBytes],
			)
			dstRow++
		}
	}
	return &RasterImage{
		Width:   img.Width,
		Height:  newHeight,
		Align:   img.Align,
		Content: newContent,
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

// WithErasePattern 返回擦除指定图案后的图像
// pattern: 要擦除的图案
// 如果图案无效（如宽度或高度小于等于0），则返回原图像
func (img *RasterImage) WithErasePattern(pattern *RasterPattern) *RasterImage {
	x, y := pattern.Search(img)
	if x < 0 || y < 0 {
		return img // 未找到匹配的图案，返回 nil
	}
	// 擦除区域
	return img.WithErase(x, y, pattern.Width, pattern.Height)
}

// WithBorder 返回添加边框后的图像
// borderWidth: 边框的宽度（单位为像素）
// 如果图像无效（如宽度或高度小于等于0，或内容为nil），则返回原图像
// 如果边框宽度小于等于0，则返回原图像
// 添加边框后，图像的宽度和高度保持不变
// 注意：边框会覆盖原图像的内容，原图像的内容将被边框覆盖
func (img *RasterImage) WithBorder(borderWidth int) *RasterImage {
	// 检查参数有效性
	if img == nil || img.Width <= 0 || img.Height <= 0 || img.Content == nil || borderWidth <= 0 {
		return img
	}

	newContent := make([]byte, len(img.Content))
	copy(newContent, img.Content)
	width := img.Width
	height := img.Height
	widthBytes := width / 8

	// 上下边框
	for row := 0; row < borderWidth; row++ {
		rowStart := row * widthBytes
		for i := 0; i < widthBytes; i++ {
			newContent[rowStart+i] = 0xFF // 全黑
		}
		rowStart = (height - 1 - row) * widthBytes
		for i := 0; i < widthBytes; i++ {
			newContent[rowStart+i] = 0xFF // 全黑
		}
	}
	// 左右边框
	for row := borderWidth; row < height-borderWidth; row++ {
		rowStart := row * widthBytes
		for col := 0; col < borderWidth; col++ {
			byteIdx := rowStart + col/8
			bitIdx := 7 - (col % 8)
			newContent[byteIdx] |= 1 << bitIdx // 左边
			byteIdx = rowStart + (width-1-col)/8
			bitIdx = 7 - ((width - 1 - col) % 8)
			newContent[byteIdx] |= 1 << bitIdx // 右边
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

// IsWhiteBorder 检查图像是否有白色边框
// 如果图像为nil或内容为nil，则返回false
func (img *RasterImage) IsWhiteBorder() bool {
	// 检查图像是否为nil或内容为nil
	if img == nil || img.Content == nil {
		return false
	}
	// 检查上边框
	for x := 0; x < img.Width; x++ {
		if img.GetPixel(x, 0) != 0 { // 如果有任何像素不是白色，则返回false
			return false
		}
	}
	// 检查下边框
	for x := 0; x < img.Width; x++ {
		if img.GetPixel(x, img.Height-1) != 0 { // 如果有任何像素不是白色，则返回false
			return false
		}
	}
	// 检查左边框
	for y := 0; y < img.Height; y++ {
		if img.GetPixel(0, y) != 0 { // 如果有任何像素不是白色，则返回false
			return false
		}
	}
	// 检查右边框
	for y := 0; y < img.Height; y++ {
		if img.GetPixel(img.Width-1, y) != 0 { // 如果有任何像素不是白色，则返回false
			return false
		}
	}
	return true // 所有边框都是白色的
}

// BlackRatio 返回图像中黑色像素的比例
// 如果图像为nil或内容为nil，则返回0
// 计算黑色像素数与总像素数的比例
// 黑色像素数是指内容中每个字节中1的个数
// 总像素数是图像宽度*高度
// 返回值范围在0到1之间，0表示全白，1表示全黑
func (img *RasterImage) BlackRatio() float64 {
	if img == nil || img.Content == nil {
		return 0 // 无效图像
	}
	blackCount := 0
	totalCount := len(img.Content) * 8 // 每个字节8位
	for _, b := range img.Content {
		blackCount += bits.OnesCount8(b) // 统计每个字节中的黑色像素数
	}
	ratio := float64(blackCount) / float64(totalCount)
	return ratio
}

// IsSingelTextLine 检查图像是否为单行文本
// 如果图像为nil或内容为nil，或高度或宽度小于等于0，则返回false
// 检查图像的顶部和底部是否有非全白的行
// 如果顶部和底部之间有非全白的行，则返回true
// 如果顶部和底部之间没有非全白的行，则返回false
// 注意：单行文本的定义是图像中只有一行非全白的内容
func (img *RasterImage) IsSingleTextLine() bool {
	if img == nil || img.Content == nil || img.Height <= 0 || img.Width <= 0 {
		return false // 无效图像
	}
	top := 0
	bottom := img.Height - 1

	// 找到第一个非全白的行
	for top < img.Height && img.IsWhiteLine(top) {
		top++
	}
	// 找到最后一个非全白的行
	for bottom >= top && img.IsWhiteLine(bottom) {
		bottom--
	}
	// 检查是否有内容
	if top > bottom {
		return false
	}
	// 检查中间每一行都不是全白
	for y := top; y <= bottom; y++ {
		if img.IsWhiteLine(y) {
			return false
		}
	}
	return true
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

// CutCharactersByConnectivity 基于连通域分析切割字符
func (img *RasterImage) CutCharacters() []*RasterImage {
	if img == nil || img.Content == nil || img.Height <= 0 || img.Width <= 0 {
		return nil // 无效图像
	}
	width, height := img.Width, img.Height
	labels := make([][]int, height)
	for i := range labels {
		labels[i] = make([]int, width)
	}
	label := 1
	type charRect struct{ minX, minY, maxX, maxY int }
	regions := map[int]*charRect{}

	// 8连通方向
	dirs := [][2]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if img.GetPixel(x, y) == 1 && labels[y][x] == 0 {
				// 新连通域
				queue := [][2]int{{x, y}}
				labels[y][x] = label
				r := &charRect{minX: x, minY: y, maxX: x, maxY: y}
				for len(queue) > 0 {
					cx, cy := queue[0][0], queue[0][1]
					queue = queue[1:]
					for _, d := range dirs {
						nx, ny := cx+d[0], cy+d[1]
						if nx >= 0 && nx < width && ny >= 0 && ny < height &&
							img.GetPixel(nx, ny) == 1 && labels[ny][nx] == 0 {
							labels[ny][nx] = label
							queue = append(queue, [2]int{nx, ny})
							// 更新外接矩形
							if nx < r.minX {
								r.minX = nx
							}
							if nx > r.maxX {
								r.maxX = nx
							}
							if ny < r.minY {
								r.minY = ny
							}
							if ny > r.maxY {
								r.maxY = ny
							}
						}
					}
				}
				regions[label] = r
				label++
			}
		}
	}

	// 提取每个连通域的外接矩形
	var characters []*RasterImage
	for _, r := range regions {
		w := r.maxX - r.minX + 1
		h := r.maxY - r.minY + 1
		if w > 0 && h > 0 {
			charImg := img.WithCrop(r.minX, r.minY, w, h)
			if charImg != nil {
				characters = append(characters, charImg)
			}
		}
	}
	return characters
}
