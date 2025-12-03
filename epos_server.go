package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/xiaohao0576/odoo-epos/raster"
)

const (
	EPOS_PULSE    = "pulse"
	EPOS_IMAGE    = "image"
	EPOS_TEST     = "test"
	EPOS_RESPONSE = `<response success="true" code="">ok</response>`
)

var ServerCert []byte
var ServerKey []byte

// ePOSHandler 处理 ePOS 打印请求
func ePOShandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h1>Hello, I am ePOS Server for Odoo</h1>"))
		return
	}
	// Set CORS headers for all requests
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	path := r.URL.Path
	suffix := "/cgi-bin/epos/service.cgi"
	if !strings.HasSuffix(path, suffix) {
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Private-Network", "true")
		w.WriteHeader(http.StatusOK)
		return
	}
	parts := strings.Split(path, "/")
	// parts[0]是空，parts[1]是name，parts[2]是"cgi-bin"，...
	if len(parts) < 5 || parts[2] != "cgi-bin" || parts[3] != "epos" || parts[4] != "service.cgi" || parts[1] == "" {
		http.NotFound(w, r)
		return
	}
	name := parts[1]
	printer, ok := Printers[name]
	if !ok {
		http.Error(w, "Printer not found", http.StatusInternalServerError)
		fmt.Println("Printer not found:", name)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	// Print the request host IP and port for debugging
	fmt.Printf("Received ePOS request from %s for printer %s\n", r.RemoteAddr, name)

	// Check the type of ePOS command, and handle accordingly
	// Check if it's an image print request
	if strings.Contains(string(body), EPOS_IMAGE) {
		img, err := raster.NewRasterImageFromXML(body)
		if err != nil {
			http.Error(w, "Failed to parse image data", http.StatusInternalServerError)
			fmt.Println("Failed to parse image data:", err)
			return
		}
		err = printer.PrintRasterImage(img)
		if err != nil {
			http.Error(w, "Failed to print image", http.StatusInternalServerError)
			fmt.Println("Failed to print image:", err)
			return
		}
		fmt.Println("Image print success.", printer)
	} else if strings.Contains(string(body), EPOS_PULSE) {
		// Check if it's a request to open the cash drawer
		err := printer.OpenCashBox()
		if err != nil {
			http.Error(w, "Failed to open cash drawer", http.StatusInternalServerError)
			fmt.Println("Failed to open cash drawer:", err)
			return
		}
		// Cash drawer opened successfully
		fmt.Println("Cash drawer opened successfully.", printer)
	} else if strings.Contains(string(body), EPOS_TEST) {
		// Handle test page print request
		err := PrintTestPage(printer)
		if err != nil {
			http.Error(w, "Failed to print test page", http.StatusInternalServerError)
			fmt.Println("Failed to print test page:", err)
			return
		}
	} else {
		fmt.Println("Unsupported ePOS command", string(body))
		http.Error(w, "Unsupported command, only support <pulse> and <image>", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(EPOS_RESPONSE))
}

func StartHttpServer() {
	LoadCertFiles()                                         // 加载证书文件
	http.HandleFunc("/eprint/png", ePrintPNGhandler)        // 处理 PNG 打印请求
	http.HandleFunc("/eprint/raw", ePrintRAWhandler)        // 处理RAW指令打印请求
	http.HandleFunc("/eprint/local", ePrintLocalPNGhandler) // 处理本地PNG文件打印请求
	http.HandleFunc("/tspl/label01", tsplhandler01)         // 处理TSPL标签打印请求
	http.HandleFunc("/tspl/label02", tsplhandler02)         // 处理TSPL标签打印请求
	http.HandleFunc("/", ePOShandler)                       // 处理根路径的请求

	cert, err := tls.X509KeyPair(ServerCert, ServerKey)
	if err != nil {
		fmt.Println("Failed to load TLS certificate:", err)
		return
	}

	hostPort := net.JoinHostPort("0.0.0.0", *Port)

	httpsServer := &http.Server{
		Addr: hostPort,
		TLSConfig: &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{cert},
		},
	}

	// 在主线程中启动 HTTPS 服务器
	fmt.Println("Version:", Version)
	fmt.Printf("Serving on https://%s\n", hostPort)
	fmt.Println("Available printers:")
	// 打印所有可用的打印机
	for name, printer := range Printers {
		fmt.Println("Printer:", name, printer)
	}
	err = httpsServer.ListenAndServeTLS("", "")
	if err != nil {
		fmt.Println("Failed to start HTTPS server:", err)
	}
}

func LoadCertFiles() {
	const certFileUrl = "https://d2ctjms1d0nxe6.cloudfront.net/cert/fullchain.pem"
	const keyFileUrl = "https://d2ctjms1d0nxe6.cloudfront.net/cert/privkey.pem"
	const certFile = "fullchain.pem"
	const keyFile = "privkey.pem"

	if fileNotExists(certFile) || fileNotExists(keyFile) || certNeedUpdate(certFile, 3) {
		fmt.Println("Certificate files need update, downloading...")
		err := DownloadFile(certFileUrl, certFile)
		if err != nil {
			fmt.Println("Failed to download certificate file:", err)
			return
		}
		err = DownloadFile(keyFileUrl, keyFile)
		if err != nil {
			fmt.Println("Failed to download key file:", err)
			return
		}
	}
	var err error
	ServerCert, err = os.ReadFile(certFile)
	if err != nil {
		fmt.Println("Failed to read certificate file:", err)
		return
	}
	ServerKey, err = os.ReadFile(keyFile)
	if err != nil {
		fmt.Println("Failed to read key file:", err)
		return
	}
}

func fileNotExists(filename string) bool {
	_, err := os.Stat(filename)
	return os.IsNotExist(err)
}

func certNeedUpdate(certFile string, thresholdDays int) bool {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return true
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return true
	}
	x509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return true
	}

	if time.Until(x509Cert.NotAfter).Hours() < float64(thresholdDays*24) {
		return true
	}
	return false
}

func DownloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: status code %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
