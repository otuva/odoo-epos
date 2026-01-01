package hwproxy

import (
	"encoding/json"
	"net/http"
)

// StatusResponse 定义状态响应的结构
type StatusResponse struct {
	Scale map[string]string `json:"scale"`
}

// StatusHandler 处理 /hw_proxy/status_json RPC POST 请求，返回硬件状态JSON数据
func (h *HwProxy) StatusHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取电子秤状态
	response := StatusResponse{
		Scale: make(map[string]string),
	}

	if h.Scale != nil {
		status := h.Scale.GetStatus()
		response.Scale["status"] = status.Status
		if status.Message != "" {
			response.Scale["message"] = status.Message
		}
	} else {
		response.Scale["status"] = "disconnected"
		response.Scale["message"] = "Scale driver not initialized"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
