package main

import (
	"flag"
	"github.com/bebo-dot-dev/go-ddns-client/service"
	"github.com/bebo-dot-dev/go-ddns-client/service/config"
	"log"
	"os"
	"time"
)

//application entry point
func main() {
	cfgFilePath := readFlags()
	cfg, ticker := config.Load(cfgFilePath)
	go service.StartServer(cfg)
	handleTicks(cfg, ticker)
}

//reads the flags (arguments) supplied to the application
func readFlags() string {
	var cfgFilePath string
	flag.StringVar(&cfgFilePath, "cfg", "", "specify the path to the serviceConfig.json file")
	flag.Parse()

	if cfgFilePath == "" {
		//unspecified config file path
		flag.Usage()
		os.Exit(1)
	}
	return cfgFilePath
}

//handles received ticks on the supplied ticker
func handleTicks(cfg *config.Configuration, ticker *time.Ticker) {
	defer ticker.Stop()
	for {
		<-ticker.C
		err := service.PerformDDNSActions(cfg)
		if err != nil {
			log.Println(err)
		}
	}
}
