package main

import (
	"flag"
)

var (
	Version    = "1.0.1"
	Printers   = make(map[string]EPrinter)
	Port       *string
	ConfigFile *string
)

func init() {
	ConfigFile = flag.String("c", "config.json", "Path to the configuration file")
	Port = flag.String("p", "443", "Port to run the server on")
	flag.Parse()
	LoadPrintersConfig(*ConfigFile)
}

func main() {
	StartHttpServer()
}
