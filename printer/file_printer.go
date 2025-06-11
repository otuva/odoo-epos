package printer

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xiaohao0576/odoo-epos/raster"
	"github.com/xiaohao0576/odoo-epos/transformer"
)

type FilePrinter struct {
	dir         string                        // 文件保存目录
	transformer transformer.RasterTransformer // 用于转换图像的转换器
}

func (p FilePrinter) String() string {
	return fmt.Sprintf("FilePrinter{Dir: %s}", p.dir)
}

func (p FilePrinter) OpenCashBox() error {
	return nil // 文件打印机不支持打开钱箱
}
func (p FilePrinter) PrintRasterImage(img *raster.RasterImage) error {
	var filename string
	if img.GetFilename() != "" {
		baseName := filepath.Base(img.GetFilename())
		filename = fmt.Sprintf("%s/%s", p.dir, baseName)
	} else {
		filename = fmt.Sprintf("%s/%s.png", p.dir, time.Now().Format("20060102-150405"))
	}
	img = p.transformer.Transform(img) // 使用转换器转换图像
	return img.SaveToPngFile(filename)
}

func (p FilePrinter) PrintRaw(data []byte) error {
	filename := fmt.Sprintf("%s/%s.bin", p.dir, time.Now().Format("20060102-150405"))
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write data to file %s: %w", filename, err)
	}
	return nil
}
