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
	tickerInterval string
)

//application entry point
func main() {
	cfgFilePath := readFlags()
	config.Load(cfgFilePath)
	startTicker()
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

//starts the application timed DNS client ticker to perform dynamic DNS updates on the configured config.AppData.UpdateInterval
func startTicker() {
	tickerInterval = config.AppData.UpdateInterval
	ticker := time.NewTicker(getTickerInterval(tickerInterval))
	defer ticker.Stop()

	go watchConfigReload(ticker)

	for {
		select {
		case _ = <-ticker.C:
			err := service.PerformDDNSActions(config.AppData)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

//watches for config.AppData reload by watching the config.AppData.Reloaded channel
func watchConfigReload(ticker *time.Ticker) {
	for {
		select {
		case _ = <-config.AppData.Reloaded:
			if config.AppData.UpdateInterval != tickerInterval {
				ticker.Reset(getTickerInterval(config.AppData.UpdateInterval))
				log.Printf("Ticker interval changed from %s to %s", tickerInterval, config.AppData.UpdateInterval)
				tickerInterval = config.AppData.UpdateInterval
			}
		}
	}
}

//parses and returns the ticker interval duration
func getTickerInterval(updateInterval string) time.Duration {
	duration, err := time.ParseDuration(updateInterval)
	if err != nil {
		//update interval parse error
		log.Panic(err)
	}
	return duration
}
