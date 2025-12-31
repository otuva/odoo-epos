package hwproxy

import (
	"net/http"
)

// HelloHandler 处理 /hw_proxy/hello GET 请求
func (h *HwProxy) HelloHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodGet {
		w.Write([]byte("ping"))
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}
