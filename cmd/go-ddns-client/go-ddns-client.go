package main

import (
	"ddns-client/service"
	"ddns-client/service/configuration"
	"flag"
	"log"
	"os"
	"time"
)

var (
	configFilename = ""
	config         = &configuration.Configuration{}
)

//application entry point
func main() {
	configFilename = readFlags()
	configuration.Load(configFilename, config)
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
	duration, err := time.ParseDuration(config.UpdateInterval)
	if err != nil {
		//update interval parse error
		log.Panic(err)
	}
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for {
		select {
		case _ = <-ticker.C:
			configuration.Load(configFilename, config)
			err = service.PerformDDNSActions(config)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
