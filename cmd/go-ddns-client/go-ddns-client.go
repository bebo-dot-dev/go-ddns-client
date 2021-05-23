package main

import (
	"flag"
	"go-ddns-client/service"
	"go-ddns-client/service/config"
	"log"
	"os"
	"time"
)

//application entry point
func main() {
	cfgFilePath := readFlags()
	config.Load(cfgFilePath)
	startDDNSTicker()
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

//starts the application timed DNS client ticker to perform dynamic DNS updates on the configured config.UpdateInterval
func startDDNSTicker() {
	ticker := time.NewTicker(getTickerInterval(config.Config.UpdateInterval))
	defer ticker.Stop()
	for {
		select {
		case _ = <-ticker.C:
			oldInterval := config.Config.UpdateInterval
			config.Load(config.Config.CfgFilePath)
			err := service.PerformDDNSActions(config.Config)
			if err != nil {
				log.Println(err)
			}
			if config.Config.UpdateInterval != oldInterval {
				ticker.Reset(getTickerInterval(config.Config.UpdateInterval))
			}
		}
	}
}

func getTickerInterval(updateInterval string) time.Duration {
	duration, err := time.ParseDuration(updateInterval)
	if err != nil {
		//update interval parse error
		log.Panic(err)
	}
	return duration
}
