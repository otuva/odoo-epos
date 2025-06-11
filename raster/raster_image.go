package raster

import (
	"fmt"
	"image"
	"image/color"
	"math/bits"
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
