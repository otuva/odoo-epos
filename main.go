package main

import (
	"flag"
	"fmt"

	eprinter "github.com/xiaohao0576/odoo-epos/printer"
)

var (
	Version    = "1.0.2"
	Port       *string
	ConfigFile *string
	Printers   eprinter.Printers
)

func init() {
	ConfigFile = flag.String("c", "config.json", "Path to the configuration file")
	Port = flag.String("p", "443", "Port to run the server on")
	flag.Parse()
	if fileNotExists(*ConfigFile) {
		fmt.Println("config file not exist, downloading...")
		const configFileUrl = "https://d2ctjms1d0nxe6.cloudfront.net/cert/config.json"
		DownloadFile(configFileUrl, *ConfigFile)
	}
}

func main() {
	Printers, _ = eprinter.LoadPrinters(*ConfigFile)
	StartHttpServer()
}
