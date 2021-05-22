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
	Notifications  Notifications          `json:"notifications"`
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
	Token        string `json:"token"`
	APIKey       string `json:"apikey"`
	APISecret    string `json:"apisecret"`
	RecordName   string `json:"recordname"`
	Port         int    `json:"port"`
	TTL          int    `json:"ttl"`
}

type Notifications struct {
	SipgateSMS SipgateSMS `json:"sipgateSMS"`
}

type SipgateSMS struct {
	Enabled   bool   `json:"enabled"`
	TokenId   string `json:"tokenId"`
	Token     string `json:"token"`
	SmsId     string `json:"smsId"`
	Recipient string `json:"recipient"`
}

// Load loads the serviceConfig.json file described by configFilename
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
