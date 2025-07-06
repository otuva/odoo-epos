// for linux usb printer support
package printer

import (
	"fmt"
	"os"
	"time"

	"github.com/xiaohao0576/odoo-epos/raster"
	"github.com/xiaohao0576/odoo-epos/transformer"
)

type USBPrinter struct {
	paperWidth        int                         // 纸张宽度
	marginBottom      int                         // 下边距
	cutCommand        []byte                      // 切纸命令
	cashDrawerCommand []byte                      // 钱箱命令
	filePath          string                      // USB打印机的文件路径
	fd                *os.File                    // 文件描述符
	transformer       transformer.TransformerFunc // 用于转换图像的转换器
}

func (p *USBPrinter) String() string {
	return fmt.Sprintf("USBPrinter{devFile: %s, paperWidth: %d, marginBottom: %d}", p.filePath, p.paperWidth, p.marginBottom)
}

func (p *USBPrinter) Open() error {
	if p.filePath == "" {
		return os.ErrInvalid
	}
	var err error
	p.fd, err = os.OpenFile(p.filePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Error opening USB printer: %v\n", err)
		return err
	}
	return nil
}

func (p *USBPrinter) OpenCashBox() error {
	err := p.Reset()
	if err != nil {
		return fmt.Errorf("failed to reset printer: %w", err)
	}
	defer p.fd.Close()
	p.fd.Write(p.cashDrawerCommand)
	return nil
}

func (p *USBPrinter) PrintRasterImage(img *raster.RasterImage) error {
	img = p.transformer(img) // 使用转换器转换图像
	if img == nil {
		return nil // 如果转换器返回 nil，表示不需要打印图像
	}
	err := p.Reset()
	if err != nil {
		return fmt.Errorf("failed to reset printer: %w", err)
	}
	defer p.fd.Close()
	for _, page := range img.CutPages() {
		page.AutoMarginLeft(p.paperWidth)
		page.AddMarginBottom(p.marginBottom)
		p.fd.Write(page.ToEscPosRasterCommand(1024))
		p.fd.Write(p.cutCommand)    // 切纸命令
		time.Sleep(1 * time.Second) // 等待打印机处理
	}

	return nil
}

func (p *USBPrinter) Reset() error {
	p.Open()
	_, err := p.fd.Write([]byte{0x1B, 0x40}) // 初始化打印机
	if err != nil {
		p.fd.Close()
		p.fd = nil
		return err
	}
	return nil
}

func (p *USBPrinter) PrintRaw(data []byte) error {
	err := p.Open()
	if err != nil {
		return fmt.Errorf("failed to open printer: %w", err)
	}
	defer p.fd.Close()
	if len(data) == 0 {
		return fmt.Errorf("no data to print")
	}
	if _, err := p.fd.Write(data); err != nil {
		return fmt.Errorf("failed to write data to printer: %w", err)
	}
	time.Sleep(1 * time.Second) // 等待打印机处理
	return nil
}
