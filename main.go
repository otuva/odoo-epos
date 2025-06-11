package main

import (
	"flag"

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
}

func main() {
	Printers, _ = eprinter.LoadPrinters(*ConfigFile)
	StartHttpServer()
}
