package service

import (
	"go-ddns-client/service/config"
	"go-ddns-client/service/ddns"
	"go-ddns-client/service/ipaddress"
	"log"
)

// PerformDDNSActions retrieves the current public IPv4 ip address and performs any json configured UpdateIPAddress actions as required
func PerformDDNSActions(cfg *config.Configuration) error {
	var err error

	if cfg.Services == nil {
		log.Println("no DDNS services configured, nothing to do")
		return nil
	}

	ipAddrProvider := getPublicIpAddressProvider(&cfg.Router)
	ipAddr, err := ipAddrProvider.GetPublicIPAddress()
	if err != nil {
		return err
	}

	if cfg.LastPublicIpAddr == nil || !ipAddr.Equal(cfg.LastPublicIpAddr) {
		for _, serviceConfig := range cfg.Services {
			ddnsClient := getDDNSClient(cfg, serviceConfig)
			if ddnsClient != nil {
				if err = ddnsClient.UpdateIPAddress(ipAddr); err != nil {
					break
				}
			}
		}
		if err == nil {
			if err = cfg.Save(ipAddr); err != nil {
				return err
			}
		}
	} else {
		log.Printf("Public IPv4 address %s remains unchanged, no DDNS updates performed", ipAddr)
	}

	return err
}

//returns an ipaddress.IAddressProvider for the supplied routerConfig *config.RouterConfiguration
func getPublicIpAddressProvider(routerConfig *config.RouterConfiguration) ipaddress.IAddressProvider {
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
func getDDNSClient(cfg *config.Configuration, serviceConfig config.ServiceConfiguration) ddns.IDynamicDnsClient {
	switch serviceConfig.ServiceType {
	case "DuckDNS":
		return ddns.DuckDNSClient{ServiceConfig: &serviceConfig, NotificationConfig: &cfg.Notifications}
	case "GoDaddy":
		return ddns.GoDaddyClient{ServiceConfig: &serviceConfig, NotificationConfig: &cfg.Notifications}
	case "Namecheap":
		return ddns.NamecheapClient{ServiceConfig: &serviceConfig, NotificationConfig: &cfg.Notifications}
	case "NoIP":
		return ddns.NoIPClient{ServiceConfig: &serviceConfig, NotificationConfig: &cfg.Notifications}
	default:
		return nil
	}
}
