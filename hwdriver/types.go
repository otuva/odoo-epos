package hwdriver

import "io"

// 定义常量表示硬件状态
const (
	StatusConnected    = "connected"
	StatusError        = "error"
	StatusConnecting   = "connecting"
	StatusDisconnected = "disconnected"
)

// HWStatus 定义硬件状态结构体
type HWStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// HWDriver 定义硬件驱动的接口
type HWDriver interface {
	GetStatus() HWStatus
}

// BaseDriver 提供硬件驱动的基础实现
type BaseDriver struct {
	dev                io.ReadWriteCloser
	DeviceIdentifier   string
	DeviceName         string
	DeviceType         string // e.g., "printer", "scanner", "scale"
	DeviceConnection   string // e.g., "usb", "serial", "network", "bluetooth", "hdmi"
	DeviceManufacturer string
}
