package config

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
)

type Configuration struct {
	CfgFilePath      string                 `json:"-"`
	UpdateInterval   string                 `json:"updateInterval"`
	LastPublicIpAddr net.IP                 `json:"lastPublicIpAddr"`
	Router           RouterConfiguration    `json:"router,omitempty"`
	Services         []ServiceConfiguration `json:"services,omitempty"`
	Notifications    Notifications          `json:"notifications,omitempty"`
}

type RouterConfiguration struct {
	RouterType   string `json:"routerType,omitempty"`
	Username     string `json:"userName,omitempty"`
	Password     string `json:"password,omitempty"`
	LoginUrl     string `json:"loginUrl,omitempty"`
	IpDetailsUrl string `json:"ipDetailsUrl,omitempty"`
}

type ServiceConfiguration struct {
	ServiceType  string `json:"serviceType"`
	TargetDomain string `json:"targetDomain"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	Token        string `json:"token,omitempty"`
	APIKey       string `json:"apikey,omitempty"`
	APISecret    string `json:"apisecret,omitempty"`
	RecordName   string `json:"recordname,omitempty"`
	Port         int    `json:"port,omitempty"`
	TTL          int    `json:"ttl,omitempty"`
}

type Notifications struct {
	SipgateSMS SipgateSMS `json:"sipgateSMS,omitempty"`
}

type SipgateSMS struct {
	Enabled   bool   `json:"enabled"`
	TokenId   string `json:"tokenId,omitempty"`
	Token     string `json:"token,omitempty"`
	SmsId     string `json:"smsId,omitempty"`
	Recipient string `json:"recipient,omitempty"`
}

var (
	Config *Configuration
)

// Load loads the serviceConfig.json file described by cfgFilePath
func Load(cfgFilePath string) *Configuration {
	jsonByteArr, err := os.ReadFile(cfgFilePath)
	if err != nil {
		//broken config file
		log.Panic(err)
	}

	err = json.Unmarshal(jsonByteArr, &Config)
	if err != nil {
		//broken json in config file
		log.Panic(err)
	}

	Config.CfgFilePath = cfgFilePath
	return Config
}

// Save persists the serviceConfig.json file to the file system along with the supplied currentPublicIpAddr
func Save(currentPublicIpAddr net.IP) error {
	if currentPublicIpAddr == nil {
		return errors.New("cannot save a nil ip address")
	}

	Config.LastPublicIpAddr = currentPublicIpAddr

	jsonByteArr, err := json.MarshalIndent(Config, "", "    ")
	if err != nil {
		return err
	}

	configFile, err := os.OpenFile(Config.CfgFilePath, os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	defer func() {
		err := configFile.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	if _, err = configFile.Write(jsonByteArr); err != nil {
		return err
	}

	return nil
}
