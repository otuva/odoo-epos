package hwproxy

import (
	"net/http"
)

// StatusHandler 处理 /hw_proxy/status_json RPC POST 请求，返回固定JSON数据
func (h *HwProxyHandler) StatusHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	status := `{"scale": {"status": "connected"}, "printer": {"status": "connected"}, "scanner": {"status": "connected"}}`
	w.Write([]byte(status))
}
