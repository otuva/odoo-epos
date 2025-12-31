package hwdriver

var Scale01 = &SerialScaleDriver{
	BaseDriver: BaseDriver{
		DeviceIdentifier:   "scale01",
		DeviceName:         "Generic Serial Scale",
		DeviceType:         "scale",
		DeviceConnection:   "serial",
		DeviceManufacturer: "Generic",
	},
	SerialProtocol: SerialProtocol{
		Port:         "/dev/ttyS0",
		BaudRate:     9600,
		ByteSize:     8,
		Parity:       0,
		StopBits:     1,
		Timeout:      5,
		WriteTimeout: 5,
	},
	MeasureRegexp:     `([+-]?\d+\.\d+)\s*(kg|g|lb)?`, // 捕获数字和小数点
	StatusRegexp:      `OK`,                           // 状态正常时的响应
	CommandTerminator: "\r\n",                         // 命令行结束符
	MeasureCommand:    "W",                            // 获取重量的命令
	EmptyAnswerValid:  false,
	RetryCount:        3,
	RetryInterval:     500,
	Debug:             false,
}
