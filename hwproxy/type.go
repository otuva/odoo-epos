package hwproxy

import (
	hwdriver "github.com/xiaohao0576/odoo-epos/hwdriver"
)

// HwProxy 用于实现 Odoo IoT Box 的 hw_proxy 相关功能
type HwProxy struct {
	Scale *hwdriver.SerialScaleDriver
}
