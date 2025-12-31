package hwproxy

import (
	"fmt"
	"net/http"
)

// StatusJSONHandler 处理 hw_proxy/scale_read jsonrpc 请求，返回电子秤的重量JSON数据
func (h *HwProxy) ScaleReadHandler(w http.ResponseWriter, r *http.Request) {
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
	var weight float64 = 1.23
	result := `{"weight": ` + fmt.Sprintf("%.2f", weight) + `}`

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(result))
}
