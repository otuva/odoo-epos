// for linux usb printer support
package printer

import (
	"fmt"
	"os"

	"github.com/xiaohao0576/odoo-epos/raster"
	"github.com/xiaohao0576/odoo-epos/transformer"
)

type USBPrinter struct {
	paperWidth        int                           // 纸张宽度
	marginBottom      int                           // 下边距
	cutCommand        []byte                        // 切纸命令
	cashDrawerCommand []byte                        // 钱箱命令
	filePath          string                        // USB打印机的文件路径
	fd                *os.File                      // 文件描述符
	transformer       transformer.RasterTransformer // 用于转换图像的转换器
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
	p.fd.Write(p.cashDrawerCommand)
	p.fd.Sync() // 确保数据写入
	return nil
}

func (p *USBPrinter) PrintRasterImage(img *raster.RasterImage) error {
	err := p.Reset()
	if err != nil {
		return fmt.Errorf("failed to reset printer: %w", err)
	}
	img = p.transformer.Transform(img) // 使用转换器转换图像
	marginLeft := img.AutoLeftMargin(p.paperWidth)
	img.AddMargin(marginLeft, p.marginBottom) // 添加边距
	p.fd.Write(img.ToEscPosRasterCommand(1024))
	p.fd.Write(p.cutCommand) // 切纸命令
	p.fd.Sync()              // 确保数据写入
	return nil
}

func (p *USBPrinter) Reset() error {
	if p.fd == nil {
		p.Open()
		if p.fd == nil {
			return fmt.Errorf("printer file descriptor is not open")
		}
	}
	_, err := p.fd.Write([]byte{0x1B, 0x40}) // 初始化打印机
	if err != nil {
		p.fd.Close()
		p.fd = nil
		return err
	}
	return nil
}

func (p *USBPrinter) PrintRaw(data []byte) error {
	err := p.Reset()
	if err != nil {
		return fmt.Errorf("failed to reset printer: %w", err)
	}
	if len(data) == 0 {
		return fmt.Errorf("no data to print")
	}
	if _, err := p.fd.Write(data); err != nil {
		return fmt.Errorf("failed to write data to printer: %w", err)
	}
	p.fd.Sync() // 确保数据写入
	return nil
}
