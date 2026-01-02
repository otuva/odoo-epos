package hwdriver

// Scale01 Generic Serial Scale - 通用串口电子秤配置
var Scale01 = SerialScaleDriver{
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
	MeasureRegexp:     `([+-]?\d+\.?\d*)\s*(kg|g|lb)?`, // 捕获数字和单位
	StatusRegexp:      `OK`,                            // 状态正常时的响应
	CommandTerminator: "\r\n",                          // 命令行结束符
	MeasureCommand:    "W",                             // 获取重量的命令
	EmptyAnswerValid:  false,
	RetryCount:        3,
	RetryInterval:     500,
	Debug:             false,
}

// ScaleYingzhan 上海英展电子秤配置
// 上海英展电子秤通常使用以下格式：
// 格式: ST,GS,+00.000kg (稳定状态，毛重，正值，重量，单位)
// 或: US,NT,+00.123kg (不稳定状态，净重，正值，重量，单位)
// 状态标志: ST=稳定(Stable), US=不稳定(Unstable)
// 类型标志: GS=毛重(Gross), NT=净重(Net), TL=去皮(Tare)
var ScaleYingzhan = SerialScaleDriver{
	BaseDriver: BaseDriver{
		DeviceIdentifier:   "yingzhan_scale",
		DeviceName:         "Shanghai Yingzhan Scale",
		DeviceType:         "scale",
		DeviceConnection:   "serial",
		DeviceManufacturer: "Yingzhan (上海英展)",
	},
	SerialProtocol: SerialProtocol{
		Port:         "/dev/ttyUSB0", // 通常USB转串口设备
		BaudRate:     9600,           // 英展常用波特率
		ByteSize:     8,
		Parity:       0, // N - No parity
		StopBits:     1, // 1个停止位
		Timeout:      2, // 2秒超时
		WriteTimeout: 2,
	},
	// 正则表达式匹配英展格式：
	// 匹配: ST,GS,+00.123kg 或 US,NT,-01.456lb
	// 捕获组: 1=状态(ST/US), 2=类型(GS/NT/TL), 3=符号(+/-), 4=重量值, 5=单位(kg/g/lb)
	MeasureRegexp:     `(?:ST|US),(?:GS|NT|TL),([+-])(\d+\.?\d*)\s*(kg|g|lb|Kg|KG|LB)`,
	StatusRegexp:      `ST`,   // 稳定状态标志
	CommandTerminator: "\r\n", // 回车换行
	CommandDelay:      0,      // 命令前延迟50ms
	MeasureDelay:      0,      // 测量后等待100ms
	NewMeasureDelay:   0,      // 新测量间隔200ms
	MeasureCommand:    "",     // 英展秤通常自动输出，不需要发送命令
	EmptyAnswerValid:  false,  // 空响应无效
	RetryCount:        5,      // 重试5次
	RetryInterval:     1000,   // 重试间隔1秒
	Debug:             true,   // 开启调试
}

// ScaleYingzhanALW 上海英展ALW系列（另一种常见协议格式）
// 格式: +00.123 kg (简化格式，直接输出符号、数值和单位)
var ScaleYingzhanALW = SerialScaleDriver{
	BaseDriver: BaseDriver{
		DeviceIdentifier:   "yingzhan_alw",
		DeviceName:         "Shanghai Yingzhan ALW Series",
		DeviceType:         "scale",
		DeviceConnection:   "serial",
		DeviceManufacturer: "Yingzhan (上海英展)",
	},
	SerialProtocol: SerialProtocol{
		Port:         "/dev/ttyUSB0",
		BaudRate:     9600,
		ByteSize:     8,
		Parity:       0,
		StopBits:     1,
		Timeout:      2,
		WriteTimeout: 2,
	},
	// 简化格式正则: +00.123 kg
	MeasureRegexp:     `([+-]?\d+\.?\d*)\s*(kg|g|lb|Kg|KG|G|LB)`,
	StatusRegexp:      `\d+\.\d+`, // 只要有数字就认为正常
	CommandTerminator: "\r\n",
	CommandDelay:      50,
	MeasureDelay:      100,
	NewMeasureDelay:   200,
	MeasureCommand:    "", // 自动输出
	EmptyAnswerValid:  false,
	RetryCount:        5,
	RetryInterval:     1000,
	Debug:             true,
}
