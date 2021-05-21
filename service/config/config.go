package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Configuration struct {
	UpdateInterval string                 `json:"updateInterval"`
	Router         RouterConfiguration    `json:"router"`
	Services       []ServiceConfiguration `json:"services"`
}

type RouterConfiguration struct {
	RouterType   string `json:"routerType"`
	Username     string `json:"userName"`
	Password     string `json:"password"`
	LoginUrl     string `json:"loginUrl"`
	IpDetailsUrl string `json:"ipDetailsUrl"`
}

type ServiceConfiguration struct {
	ServiceType  string `json:"serviceType"`
	TargetDomain string `json:"targetDomain"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

//loads the serviceConfig.json file described by configFilename
func Load(configFilename string, config *Configuration) {
	jsonByteArr, err := ioutil.ReadFile(configFilename)
	if err != nil {
		//broken config file
		log.Panic(err)
	}

	err = json.Unmarshal(jsonByteArr, &config)
	if err != nil {
		//broken json in config file
		log.Panic(err)
	}
}
