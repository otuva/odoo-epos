package hwproxy

import (
	"encoding/json"
	"net/http"
)

// ScaleReadResponse 定义称重响应的结构
type ScaleReadResponse struct {
	Weight float64 `json:"weight"`
}

// ScaleReadHandler 处理 hw_proxy/scale_read jsonrpc 请求，返回电子秤的重量JSON数据
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

	// 读取电子秤重量
	var weight float64 = 0.0

	if h.Scale != nil {
		// 调用ReadWeight来获取最新重量
		err := h.Scale.ReadWeight()
		if err == nil {
			weight = h.Scale.GetWeight()
		}
		// 即使读取失败，也返回上次的重量值
	}

	response := ScaleReadResponse{
		Weight: weight,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
