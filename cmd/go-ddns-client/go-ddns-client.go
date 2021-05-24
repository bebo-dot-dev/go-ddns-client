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
	tickerInterval = config.AppData.UpdateInterval
	ticker := time.NewTicker(getTickerInterval(tickerInterval))
	defer ticker.Stop()
	for {
		select {
		case _ = <-ticker.C:
			err := service.PerformDDNSActions(config.AppData)
			if err != nil {
				log.Println(err)
			}
			checkTickerInterval(ticker)
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

func checkTickerInterval(ticker *time.Ticker) {
	if config.AppData.UpdateInterval != tickerInterval {
		ticker.Reset(getTickerInterval(config.AppData.UpdateInterval))
		tickerInterval = config.AppData.UpdateInterval
	}
}
