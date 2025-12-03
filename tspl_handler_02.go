package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// Label02 定义了一个标签的结构体，用于咖啡奶茶店的标签打印
// 包含名称、条形码、价格、备注等字段
// 该结构体用于生成 TSPL 格式的标签打印指令
type label02 struct {
	CompanyName       string  `json:"company_name"`
	TableNumber       int     `json:"table_number"`
	TrackingNumber    string  `json:"tracking_number"`
	CustomerName      string  `json:"customer_name"`
	ProductName       string  `json:"product_name"`
	ProductAttributes string  `json:"product_attributes"`
	Notes             string  `json:"notes"`
	DefaultCode       string  `json:"default_code"`
	Barcode           string  `json:"barcode"`
	PriceUnit         float64 `json:"price_unit"`
	CurrencyCode      string  `json:"currency_code"`
	Page              string  `json:"page"`
	FullProductName   string  `json:"full_product_name"`
	Qty               int     `json:"qty"`
	CurrencySymbol    string  `json:"currency_symbol"`
	CurrencyName      string  `json:"currency_name"`
}
type Label02List []label02

// encodeToGB18030 将字符串转换为GB18030编码
// 用于所有需要发送给打印机的中文字符串
func encodeToGB18030String(text string) string {
	encoder := simplifiedchinese.GB18030.NewEncoder()
	encodedBytes, _ := encoder.Bytes([]byte(text))
	return string(encodedBytes)
}

// ToTSPL 方法将单个 Label02 转换为 TSPL 内容指令（不包含打印机配置）
// 这个方法主要供 Label02List.ToTSPL() 调用，用于生成单个标签的内容部分
// 返回标签内容的 TSPL 指令切片，不包含 SIZE、GAP 等配置命令和 PRINT 命令
// 在XPrinter XP-236B打印机上，标签宽度为40mm，高度为30mm,测试通过
func (label label02) toTSPL() []string {

	// 创建标签内容命令（不包含配置和打印命令）
	tsplCommands := []string{}

	// 打印tracking number
	tsplCommands = append(tsplCommands, fmt.Sprintf("TEXT 10,10,\"TSS24.BF2\",0,1,1,\"%s%s\"", "#", encodeToGB18030String(label.TrackingNumber)))

	//打印桌号
	if label.TableNumber != 0 {
		tableNumberStr := fmt.Sprintf("%d", label.TableNumber)
		tsplCommands = append(tsplCommands, fmt.Sprintf("TEXT 90,10,\"TSS24.BF2\",0,1,1,\"%s%s\"", "Table:", encodeToGB18030String(tableNumberStr)))
	}

	// 在右上角打印当前页码
	tsplCommands = append(tsplCommands, fmt.Sprintf("TEXT 230,10,\"TSS24.BF2\",0,1,1,\"%s\"", encodeToGB18030String(label.Page)))

	// 打印产品名称
	tsplCommands = append(tsplCommands, fmt.Sprintf("TEXT 10,40,\"TSS24.BF2\",0,1,2,\"%s\"", encodeToGB18030String(label.ProductName)))

	// 打印产品属性
	tsplCommands = append(tsplCommands, fmt.Sprintf("TEXT 10,100,\"TSS24.BF2\",0,1,1,\"%s\"", encodeToGB18030String(label.ProductAttributes)))

	// 打印备注
	tsplCommands = append(tsplCommands, fmt.Sprintf("TEXT 10,150,\"TSS24.BF2\",0,1,1,\"%s\"", encodeToGB18030String(label.Notes)))

	// 打印价格
	var priceText string = fmt.Sprintf("%.2f", label.PriceUnit)
	if label.CurrencyCode == "USD" {
		priceText = fmt.Sprintf("%s%.2f", "$", label.PriceUnit)
	}
	if label.CurrencyCode == "KHR" {
		priceText = fmt.Sprintf("%.0f %s", label.PriceUnit, "R")
	}
	tsplCommands = append(tsplCommands, fmt.Sprintf("TEXT 10,185,\"TSS24.BF2\",0,2,2,\"%s\"", encodeToGB18030String(priceText)))

	// 打印时间，格式 HH:MM
	timeString := time.Now().Format("15:04")
	tsplCommands = append(tsplCommands, fmt.Sprintf("TEXT 230,210,\"TSS24.BF2\",0,1,1,\"%s\"", encodeToGB18030String(timeString)))

	return tsplCommands
}

// ToTSPL 方法将 Label02List 转换为 TSPL 格式的打印指令
// 返回一个字节切片，包含所有标签的 TSPL 指令
// 对于批量打印，只设置一次打印机配置，然后为每个标签设置内容
func (labels Label02List) ToTSPL() []byte {
	if len(labels) == 0 {
		return []byte{}
	}

	// 创建TSPL命令
	tsplCommands := []string{}

	// 只在开头设置一次打印机配置
	tsplCommands = append(tsplCommands, "SIZE 40 mm, 30 mm")
	tsplCommands = append(tsplCommands, "GAP 2 mm, 0")
	tsplCommands = append(tsplCommands, "DIRECTION 1")
	tsplCommands = append(tsplCommands, "REFERENCE 0,0")
	tsplCommands = append(tsplCommands, "OFFSET 0 mm")
	tsplCommands = append(tsplCommands, "SET PEEL OFF")
	tsplCommands = append(tsplCommands, "SET CUTTER OFF")
	tsplCommands = append(tsplCommands, "SET PARTIAL_CUTTER OFF")
	tsplCommands = append(tsplCommands, "SET TEAR ON")
	tsplCommands = append(tsplCommands, "DENSITY 8") // 设置打印浓度
	tsplCommands = append(tsplCommands, "SPEED 4")   // 设置打印速度
	tsplCommands = append(tsplCommands, "CLS")

	// 为每个标签添加内容
	for _, label := range labels {
		// 调用单个标签的ToTSPL方法获取内容指令
		labelCommands := label.toTSPL()
		tsplCommands = append(tsplCommands, labelCommands...)

		// 计算该标签的打印份数
		copies := 1 // 默认打印1份

		// 一次性打印该标签的所有副本，而不是循环多次
		tsplCommands = append(tsplCommands, fmt.Sprintf("PRINT %d,1", copies))

		// 如果有多个不同的标签，需要清除缓冲区准备下一个标签
		if len(labels) > 1 {
			tsplCommands = append(tsplCommands, "CLS")
		}
	}

	// 将所有命令连接成字节切片，使用\r\n作为行分隔符
	result := ""
	for _, cmd := range tsplCommands {
		result += cmd + "\r\n"
	}

	return []byte(result)
}

func tsplhandler02(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Private-Network", "true")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"success":false,"msg":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var labels Label02List
	var printerName = "xp-236b" // 默认打印机名称
	// 解析请求体中的标签数据

	if err := json.NewDecoder(r.Body).Decode(&labels); err != nil {
		http.Error(w, `{"success":false,"msg":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	if len(labels) == 0 {
		http.Error(w, `{"success":false,"msg":"No labels provided"}`, http.StatusBadRequest)
		return
	}
	// 将标签转换为TSPL格式
	tsplData := labels.ToTSPL()

	// 获取打印机名称
	printer, ok := Printers[printerName]
	if !ok {
		http.Error(w, `{"success":false,"msg":"Printer not found"}`, http.StatusBadRequest)
		return
	}
	// 打印TSPL数据
	if err := printer.PrintRaw(tsplData); err != nil {
		http.Error(w, `{"success":false,"msg":"Failed to print raw commands: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true,"msg":"Labels printed successfully"}`))
}
