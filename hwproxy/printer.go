package hwproxy

import (
	"net/http"
)

// StatusJSONHandler 处理 /hw_proxy/default_printer_action jsonrpc 请求，返回固定JSON数据
func (h *HwProxyHandler) DefaultPrinterActionHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// do something here, currently always return true
	result := `true`

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(result))
}
