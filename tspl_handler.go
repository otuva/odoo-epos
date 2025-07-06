package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// Label01 定义了一个标签的结构体
// 包含名称、条形码、价格、备注和副本数量等字段
// 该结构体用于生成 TSPL 格式的标签打印指令
type label01 struct {
	Name    string  `json:"name"`
	Barcode string  `json:"barcode"`
	Price   float64 `json:"price"`
	Remark  string  `json:"remark"`
	Copies  int     `json:"copies"`
}
type Label01List []label01

// ToTSPL 方法将单个 Label01 转换为 TSPL 内容指令（不包含打印机配置）
// 这个方法主要供 Label01List.ToTSPL() 调用，用于生成单个标签的内容部分
// 返回标签内容的 TSPL 指令切片，不包含 SIZE、GAP 等配置命令和 PRINT 命令
func (label label01) toTSPL() []string {
	// 将中文文本转换为GB18030编码
	encoder := simplifiedchinese.GB18030.NewEncoder()
	nameBytes, _ := encoder.Bytes([]byte(label.Name))
	remarkBytes, _ := encoder.Bytes([]byte(label.Remark))

	// 创建标签内容命令（不包含配置和打印命令）
	tsplCommands := []string{}

	// 产品名称 (最上方)
	tsplCommands = append(tsplCommands, fmt.Sprintf("TEXT 10,10,\"TSS24.BF2\",0,1,1,\"%s\"", string(nameBytes)))

	// EAN13条形码 (中间)
	if len(label.Barcode) == 13 {
		tsplCommands = append(tsplCommands, fmt.Sprintf("BARCODE 10,60,\"EAN13\",60,1,0,2,2,\"%s\"", label.Barcode))
	}

	// 价格 (下方)
	priceText := fmt.Sprintf("¥%.2f", label.Price)
	tsplCommands = append(tsplCommands, fmt.Sprintf("TEXT 10,140,\"TSS24.BF2\",0,1,1,\"%s\"", priceText))

	// 备注 (最下方)
	if label.Remark != "" {
		tsplCommands = append(tsplCommands, fmt.Sprintf("TEXT 10,180,\"TSS24.BF2\",0,1,1,\"%s\"", string(remarkBytes)))
	}

	return tsplCommands
}

// ToTSPL 方法将 Label01List 转换为 TSPL 格式的打印指令
// 返回一个字节切片，包含所有标签的 TSPL 指令
// 对于批量打印，只设置一次打印机配置，然后为每个标签设置内容
func (labels Label01List) ToTSPL() []byte {
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
	tsplCommands = append(tsplCommands, "CLS")

	// 为每个标签添加内容
	for _, label := range labels {
		// 调用单个标签的ToTSPL方法获取内容指令
		labelCommands := label.toTSPL()
		tsplCommands = append(tsplCommands, labelCommands...)

		// 计算该标签的打印份数
		copies := label.Copies
		if copies <= 0 {
			copies = 1
		}

		// 为每一份添加打印命令
		for i := 0; i < copies; i++ {
			tsplCommands = append(tsplCommands, "PRINT 1,1")
		}
	}

	// 将所有命令连接成字节切片，使用\r\n作为行分隔符
	result := ""
	for _, cmd := range tsplCommands {
		result += cmd + "\r\n"
	}

	return []byte(result)
}

func Tsplhandler01(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"success":false,"msg":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var labels Label01List
	var printerName = "tspl_printer" // 默认打印机名称
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
