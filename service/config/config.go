package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type Configuration struct {
	CfgFilePath        string                 `json:"-"`
	Reloaded           chan bool              `json:"-"`              // A channel upon which config reload events are delivered
	LastUpdateInterval string                 `json:"-"`              // Used to track changes to the update interval
	FileInfo           os.FileInfo            `json:"-"`              // Used to track changes to the config file
	Mu                 *sync.Mutex            `json:"-"`              // Used to lock and unlock access to the package level cfg
	UpdateInterval     string                 `json:"updateInterval"` // A duration string parsed by time.ParseDuration
	LastPublicIpAddr   net.IP                 `json:"lastPublicIpAddr"`
	Router             RouterConfiguration    `json:"router,omitempty"`
	Services           []ServiceConfiguration `json:"services,omitempty"`
	Notifications      Notifications          `json:"notifications,omitempty"`
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
	Email      Email      `json:"email,omitempty"`
}

type SipgateSMS struct {
	Enabled   bool   `json:"enabled"`
	TokenId   string `json:"tokenId,omitempty"`
	Token     string `json:"token,omitempty"`
	SmsId     string `json:"smsId,omitempty"`
	Recipient string `json:"recipient,omitempty"`
}

type Email struct {
	IsEnabled    bool   `json:"enabled"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	From         EmailAddress
	Recipients   []EmailAddress
	SmtpServer   string `json:"smtpServer,omitempty"`
	SecurityType string `json:"securityType,omitempty"` /*SSL or TLS*/
}

type EmailAddress struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address,omitempty"`
}

var (
	cfg *Configuration //the package level cfg
)

//unmarshalConfigFile performs a json.Unmarshal call on the supplied cfgFilePath to deserialize the cfg file to the
//package level cfg *Configuration variable
func unmarshalConfigFile(cfgFilePath string) {
	jsonByteArr, err := os.ReadFile(cfgFilePath)
	if err != nil {
		//broken config file
		log.Panic(err)
	}

	err = json.Unmarshal(jsonByteArr, &cfg)
	if err != nil {
		//broken json in config file
		log.Panic(err)
	}

	cfg.FileInfo, err = os.Stat(cfgFilePath)
	if err != nil {
		log.Panic(err)
	}
}

//watchConfigFile implements a simple file watcher on the cfg.cfgFilePath file to enable reload on change detection
func (appData *Configuration) watchConfigFile() {
	for {
		nowFileInfo, err := os.Stat(appData.CfgFilePath)
		if err != nil {
			log.Panic(err)
		}

		if nowFileInfo.Size() != appData.FileInfo.Size() || nowFileInfo.ModTime() != appData.FileInfo.ModTime() {
			//refresh on change
			appData.Mu.Lock()
			unmarshalConfigFile(appData.CfgFilePath)
			appData.FileInfo = nowFileInfo
			appData.Mu.Unlock()

			log.Printf("A change was detected on %s, the file was reloaded", appData.CfgFilePath)
			appData.Reloaded <- true //channel comm
		}
		time.Sleep(1 * time.Second)
	}
}

//parses and returns the ticker interval duration
func (appData *Configuration) getTickerInterval(updateInterval string) time.Duration {
	duration, err := time.ParseDuration(updateInterval)
	if err != nil {
		//update interval parse error
		log.Panic(err)
	}
	return duration
}

//handles appData reload by listening on the appData.reloaded channel
func (appData *Configuration) handleConfigReload(ticker *time.Ticker) {
	for {
		select {
		case _ = <-appData.Reloaded:
			if appData.UpdateInterval != appData.LastUpdateInterval {
				//interval change ticker reset
				ticker.Reset(appData.getTickerInterval(appData.UpdateInterval))
				log.Printf("**Ticker interval changed from %s to %s**", appData.LastUpdateInterval, appData.UpdateInterval)
				appData.LastUpdateInterval = appData.UpdateInterval
			}
		}
	}
}

//creates a new ticker to perform ddns updates on the configured appData.UpdateInterval
func (appData *Configuration) createTicker() *time.Ticker {
	appData.LastUpdateInterval = appData.UpdateInterval
	ticker := time.NewTicker(appData.getTickerInterval(appData.LastUpdateInterval))
	log.Printf("Ticker created with a %s interval", appData.LastUpdateInterval)

	go appData.handleConfigReload(ticker)

	return ticker
}

/*
Load loads the serviceConfig.json file described by cfgFilePath and sets up a config file watcher to detect changes to
enable config reload on change. Load also creates a new Time.Ticker with a tick duration set to the currently
configured appData.UpdateInterval
*/
func Load(cfgFilePath string) (*Configuration, *time.Ticker) {
	unmarshalConfigFile(cfgFilePath)

	cfg.CfgFilePath = cfgFilePath
	cfg.Mu = &sync.Mutex{}
	cfg.Reloaded = make(chan bool)

	go cfg.watchConfigFile() //spin the file watcher into go routine

	ticker := cfg.createTicker()

	return cfg, ticker
}

// Save persists the serviceConfig.json file to the file system with the supplied currentPublicIpAddr
func (appData *Configuration) Save(currentPublicIpAddr net.IP) error {
	if currentPublicIpAddr == nil {
		return errors.New("cannot save a nil ip address")
	}

	appData.LastPublicIpAddr = currentPublicIpAddr

	jsonByteArr, err := json.MarshalIndent(appData, "", "    ")
	if err != nil {
		return err
	}

	configFile, err := os.OpenFile(appData.CfgFilePath, os.O_RDWR|os.O_TRUNC, 0600)
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

//GetDomainsStr returns a comma separated string of all configured target domain names
func (appData *Configuration) GetDomainsStr() (string, error) {
	var err error
	var builder strings.Builder
	for index, svc := range cfg.Services {
		_, err = fmt.Fprintf(&builder, "%s", svc.TargetDomain)
		if err != nil {
			return "", err
		}
		if index < (len(cfg.Services) - 1) {
			_, err = fmt.Fprint(&builder, ",")
			if err != nil {
				return "", err
			}
		}
	}
	return builder.String(), err
}
