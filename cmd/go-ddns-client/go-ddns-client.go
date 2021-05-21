package main

import (
	"flag"
	"go-ddns-client/service"
	"go-ddns-client/service/config"
	"log"
	"os"
	"time"
)

var (
	configFilename = ""
	cfg         = &config.Configuration{}
)

//application entry point
func main() {
	configFilename = readFlags()
	config.Load(configFilename, cfg)
	startDynamicDnsClientTicker()
}

//reads the flags (arguments) supplied to the application
func readFlags() string {
	var configFilename string
	flag.StringVar(&configFilename, "cfg", "", "specify the path to the serviceConfig.json file")
	flag.Parse()

	if configFilename == "" {
		//unspecified config file path
		flag.Usage()
		os.Exit(1)
	}
	return configFilename
}

//starts the application timed DNS client ticker to perform dynamic DNS updates on the configured config.UpdateInterval
func startDynamicDnsClientTicker() {
	duration, err := time.ParseDuration(cfg.UpdateInterval)
	if err != nil {
		//update interval parse error
		log.Panic(err)
	}
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for {
		select {
		case _ = <-ticker.C:
			config.Load(configFilename, cfg)
			err = service.PerformDDNSActions(cfg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
