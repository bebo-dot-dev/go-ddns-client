package config

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Configuration struct {
	CfgFilePath      string                 `json:"-"`
	Reloaded         chan bool              `json:"-"` // A channel upon which config reload events are delivered
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
	AppData        *Configuration
	mu             = &sync.Mutex{}
	configFileInfo os.FileInfo
)

//unmarshalConfigFile performs a json.Unmarshal call on the supplied cfgFilePath to deserialize the file to the
//package level AppData *Configuration variable
func unmarshalConfigFile(cfgFilePath string) {
	jsonByteArr, err := os.ReadFile(cfgFilePath)
	if err != nil {
		//broken config file
		log.Panic(err)
	}

	err = json.Unmarshal(jsonByteArr, &AppData)
	if err != nil {
		//broken json in config file
		log.Panic(err)
	}
}

//watchConfigFile implements a simple file watcher on the AppData.CfgFilePath file to enable reload on change detection
func watchConfigFile() {
	if AppData != nil {
		for {
			nowFileInfo, err := os.Stat(AppData.CfgFilePath)
			if err != nil {
				log.Panic(err)
			}

			if nowFileInfo.Size() != configFileInfo.Size() || nowFileInfo.ModTime() != configFileInfo.ModTime() {
				//refresh on change
				mu.Lock()
				unmarshalConfigFile(AppData.CfgFilePath)
				configFileInfo = nowFileInfo
				mu.Unlock()
				log.Printf("A change was detected on %s, the file was reloaded", AppData.CfgFilePath)
				AppData.Reloaded <- true
			}
			time.Sleep(1 * time.Second)
		}
	}
}

// Load loads the serviceConfig.json file described by cfgFilePath
func Load(cfgFilePath string) {
	var err error
	if AppData == nil {
		configFileInfo, err = os.Stat(cfgFilePath)
		if err != nil {
			log.Panic(err)
		}
		unmarshalConfigFile(cfgFilePath)
		AppData.CfgFilePath = cfgFilePath
		AppData.Reloaded = make(chan bool)
		go watchConfigFile() //spin the file watcher into go routine
	}
}

// Save persists the serviceConfig.json file to the file system along with the supplied currentPublicIpAddr
func Save(currentPublicIpAddr net.IP) error {
	if currentPublicIpAddr == nil {
		return errors.New("cannot save a nil ip address")
	}

	AppData.LastPublicIpAddr = currentPublicIpAddr

	jsonByteArr, err := json.MarshalIndent(AppData, "", "    ")
	if err != nil {
		return err
	}

	configFile, err := os.OpenFile(AppData.CfgFilePath, os.O_RDWR|os.O_TRUNC, 0755)
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
