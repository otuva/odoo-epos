package printer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/xiaohao0576/odoo-epos/raster"
	"github.com/xiaohao0576/odoo-epos/transformer"
)

type EPrinter interface {
	OpenCashBox() error
	PrintRasterImage(img *raster.RasterImage) error
	PrintRaw(data []byte) error
}

type Printers = map[string]EPrinter

type ConfigPrinter struct {
	Type              string `json:"type"`                // 打印机类型
	Address           string `json:"address"`             // 打印机地址
	PaperWidth        int    `json:"paper_width"`         // 纸张宽度
	MarginBottom      int    `json:"margin_bottom"`       // 下边距
	CutCommnad        string `json:"cut_command"`         // 切纸命令
	CashDrawerCommand string `json:"cash_drawer_command"` // 钱箱命令
	Transformer       string `json:"transformer"`         // 图像转换器
}

func (c *ConfigPrinter) NewPrinter() EPrinter {
	//设置默认值和边距
	if c.PaperWidth <= 0 {
		c.PaperWidth = 576 // 默认纸张宽度
	}

	if c.MarginBottom <= 0 {
		c.MarginBottom = 120 // 默认下边距
	}

	cutCommand, err := hex.DecodeString(c.CutCommnad)
	if err != nil || len(cutCommand) == 0 {
		cutCommand = []byte{0x1D, 0x56, 0x01} // 默认半切纸命令
	}
	cashDrawerCommand, err := hex.DecodeString(c.CashDrawerCommand)
	if err != nil || len(cashDrawerCommand) == 0 {
		cashDrawerCommand = []byte{0x1B, 0x70, 0x00, 0x19, 0xFA} // 默认钱箱命令
	}

	transfer, ok := transformer.Transformers[c.Transformer]
	if !ok {
		transfer = transformer.Identity // 使用默认转换器
	}

	switch c.Type {
	case "usb":
		return &USBPrinter{
			paperWidth:        c.PaperWidth,      // 纸张宽度
			marginBottom:      c.MarginBottom,    // 下边距
			filePath:          c.Address,         // USB打印机的文件路径
			cutCommand:        cutCommand,        // 切纸命令
			cashDrawerCommand: cashDrawerCommand, // 钱箱命令
			transformer:       transfer,          // 图像转换器
		}
	case "tcp":
		return &TCPPrinter{
			paperWidth:        c.PaperWidth,      // 纸张宽度
			marginBottom:      c.MarginBottom,    // 下边距
			HostPort:          c.Address,         // 打印机地址
			cutCommand:        cutCommand,        // 切纸命令
			cashDrawerCommand: cashDrawerCommand, // 钱箱命令
			transformer:       transfer,          // 图像转换器
		}
	case "serial":
		return &SerialPrinter{
			paperWidth:        c.PaperWidth,      // 纸张宽度
			marginBottom:      c.MarginBottom,    // 下边距
			serialConfig:      c.Address,         // 串口地址
			cutCommand:        cutCommand,        // 切纸命令
			cashDrawerCommand: cashDrawerCommand, // 钱箱命令
			transformer:       transfer,          // 图像转换器
		}
	case "file":
		return &FilePrinter{
			dir:         c.Address, // 文件保存目录
			transformer: transfer,  // 图像转换器
		}
	default:
		return nil // 未知类型
	}
}

// 读取并解析 config.json 到 Printers
func LoadPrinters(filename string) (Printers, error) {
	var printers Printers = make(map[string]EPrinter)
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening config file: %v\n", err)
		return nil, err
	}
	defer file.Close()

	configPrinters := make(map[string]ConfigPrinter)
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&configPrinters); err != nil {
		fmt.Printf("Error decoding config file: %v\n", err)
		return nil, err
	}
	for name, config := range configPrinters {
		printer := config.NewPrinter()
		if printer == nil {
			fmt.Printf("Unknown printer type for %s: %s\n", name, config.Type)
			continue
		}
		printers[name] = printer
	}
	if len(printers) == 0 {
		fmt.Println("No printers configured")
		return nil, fmt.Errorf("no printers configured or all failed to open")
	}
	return printers, nil
}
