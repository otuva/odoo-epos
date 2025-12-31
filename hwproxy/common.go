package hwproxy

import "net/http"

// setCORSHeaders 设置通用CORS头
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Private-Network", "true")
}

func (h *HwProxy) NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/hw_proxy/hello", h.HelloHandler)
	mux.HandleFunc("/hw_proxy/status_json", h.StatusHandler)
	mux.HandleFunc("/hw_proxy/scale_read", h.ScaleReadHandler)
	mux.HandleFunc("/hw_proxy/default_printer_action", h.DefaultPrinterActionHandler)
	return mux
}
