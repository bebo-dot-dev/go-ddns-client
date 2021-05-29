package service

import (
	"fmt"
	"go-ddns-client/service/config"
	"go-ddns-client/service/ddns"
	"go-ddns-client/service/ipaddress"
	"go-ddns-client/service/notifications"
	"io"
	"log"
	"net"
	"net/http"
)

//StartServer starts a http server on port 8080 to serve up the current ipv4 and ipv6 ip addresses
func StartServer(cfg *config.Configuration) {
	ipv4Handler := func(w http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(w, cfg.LastIPv4.String())
		if err != nil {
			log.Printf("io.WriteString error in ipv4Handler: %v", err)
		}
	}
	ipv6Handler := func(w http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(w, cfg.LastIPv6.String())
		if err != nil {
			log.Printf("io.WriteString error in ipv6Handler: %v", err)
		}
	}
	jsonHandler := func(w http.ResponseWriter, req *http.Request) {
		const json = `{
	"hostname": "%s",
	"ipv4": "%s",
	"ipv6": "%s"
}`
		w.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(w, fmt.Sprintf(json, cfg.Hostname, cfg.LastIPv4, cfg.LastIPv6))
		if err != nil {
			log.Printf("io.WriteString error in jsonHandler: %v", err)
		}
	}

	http.HandleFunc("/ipv4", ipv4Handler)
	http.HandleFunc("/ipv6", ipv6Handler)
	http.HandleFunc("/json", jsonHandler)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, nil))
}

//PerformDDNSActions retrieves the current public IPv4 ip address and performs any json configured UpdateIPAddresses actions as required
func PerformDDNSActions(cfg *config.Configuration) error {
	var err error

	if cfg.Services == nil {
		log.Println("no DDNS services configured, nothing to do")
		return nil
	}

	ipAddrProvider := getIpAddressProvider(&cfg.Router)
	ipv4, ipv6, err := ipAddrProvider.GetPublicIPAddresses()
	if err != nil {
		return err
	}

	if cfg.IPAddressesChanged(ipv4, ipv6) {
		for _, serviceConfig := range cfg.Services {
			ddnsClient := getDDNSClient(&serviceConfig)
			if ddnsClient != nil {
				if err = ddnsClient.UpdateIPAddresses(ipv4, ipv6); err != nil {
					break
				}
			}
		}
		if err == nil {
			if err = cfg.Save(ipv4, ipv6); err != nil {
				return err
			}
			if err = sendNotifications(cfg, ipv4, ipv6); err != nil {
				return err
			}
		}
	} else {
		log.Printf("IPv4 address %s and IPv6 %s remain unchanged, no DDNS updates performed", ipv4, ipv6)
	}

	return err
}

//getIpAddressProvider returns an ipaddress.IAddressProvider for the supplied routerConfig *config.RouterConfiguration
func getIpAddressProvider(routerConfig *config.RouterConfiguration) ipaddress.IAddressProvider {
	var ipAddressProvider ipaddress.IAddressProvider
	if routerConfig != nil && routerConfig.RouterType != "" {
		switch routerConfig.RouterType {
		case "BTSmartHub2":
			ipAddressProvider = &ipaddress.BTSmartHub2{Config: routerConfig}
		}
	} else {
		ipAddressProvider = &ipaddress.Default{}
	}
	return ipAddressProvider
}

//returns the corresponding ddns.IDynamicDnsClient for the supplied serviceConfig.ServiceType
func getDDNSClient(serviceConfig *config.ServiceConfiguration) ddns.IDynamicDnsClient {
	switch serviceConfig.ServiceType {
	case "DuckDNS":
		return ddns.DuckDNSClient{ServiceConfig: serviceConfig}
	case "GoDaddy":
		return ddns.GoDaddyClient{ServiceConfig: serviceConfig}
	case "Namecheap":
		return ddns.NamecheapClient{ServiceConfig: serviceConfig}
	case "NoIP":
		return ddns.NoIPClient{ServiceConfig: serviceConfig}
	default:
		return nil
	}
}

//sendNotifications sends all configured notifications on ip address change
func sendNotifications(cfg *config.Configuration, ipv4, ipv6 net.IP) error {
	var err error
	mgr := notifications.GetManager(&cfg.Notifications)
	if mgr.GetNotifierCount() > 0 {
		domainsStr, err := cfg.GetDomainsStr()
		if err != nil {
			return err
		}
		err = mgr.Send(cfg.Hostname, len(cfg.Services), domainsStr, ipv4.String(), ipv6.String())
	}
	return err
}
