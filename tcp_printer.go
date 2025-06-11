package main

import (
	"fmt"
	"net"
	"time"
)

type TCPPrinter struct {
	paperWidth        int               // 纸张宽度
	marginBottom      int               // 下边距
	cutCommand        []byte            //切纸命令
	cashDrawerCommand []byte            // 钱箱命令
	HostPort          string            // 打印机地址
	fd                net.Conn          // 直接用 net.Conn
	transformer       RasterTransformer // 用于转换图像的转换器
}

func (p *TCPPrinter) String() string {
	return fmt.Sprintf("TCPPrinter{HostPort: %s, paperWidth: %d, marginBottom: %d}", p.HostPort, p.paperWidth, p.marginBottom)
}

func (p *TCPPrinter) Open() error {
	if p.HostPort == "" {
		return net.ErrClosed
	}
	if p.fd != nil {
		p.fd.Close()
	}
	conn, err := net.DialTimeout("tcp", p.HostPort, 5*time.Second)
	if err != nil {
		return err
	}
	p.fd = conn
	p.fd.SetWriteDeadline(time.Now().Add(5 * time.Second)) // 设置写超时
	// 发送初始化命令
	_, err = p.fd.Write([]byte{0x1B, 0x40}) // 初始化打印机
	if err != nil {
		p.fd.Close() // 如果初始化失败，关闭连接
		p.fd = nil
		return err
	}
	return nil
}

func (p *TCPPrinter) Close() error {
	if p.fd != nil {
		err := p.fd.Close()
		p.fd = nil
		return err
	}
	return nil
}

func (p *TCPPrinter) OpenCashBox() error {
	if p.fd == nil {
		if err := p.Open(); err != nil {
			return err
		}
	}
	defer p.Close() // 确保在函数结束时关闭连接
	// 发送打开钱箱的命令
	_, err := p.fd.Write(p.cashDrawerCommand)
	return err
}

func (p *TCPPrinter) PrintRasterImage(img *RasterImage) error {
	if p.fd == nil {
		if err := p.Open(); err != nil {
			return err
		}
	}
	defer p.Close()
	img = p.transformer.Transform(img) // 使用转换器转换图像
	marginLeft := img.AutoLeftMargin(p.paperWidth)
	img.AddMargin(marginLeft, p.marginBottom)
	if _, err := p.fd.Write(img.ToEscPosRasterCommand(1024)); err != nil {
		return err
	}
	if _, err := p.fd.Write(p.cutCommand); err != nil {
		return err
	}
	return nil
}

func (p *TCPPrinter) PrintRaw(data []byte) error {
	if p.fd == nil {
		if err := p.Open(); err != nil {
			return err
		}
	}
	defer p.Close()
	if len(data) == 0 {
		return fmt.Errorf("no data to print")
	}
	if _, err := p.fd.Write(data); err != nil {
		return fmt.Errorf("failed to write data to printer: %w", err)
	}
	return nil
}
