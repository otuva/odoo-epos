package main

import (
	"flag"
)

var (
	Version    = "dev"
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
