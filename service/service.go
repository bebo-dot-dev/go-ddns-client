package service

import (
	"go-ddns-client/service/config"
	"go-ddns-client/service/ddns"
	"go-ddns-client/service/ipaddress"
	"log"
	"net"
)

var (
	currentPublicIpAddr net.IP
	lastPublicIpAddr    net.IP
)

//retrieves the current public IPv4 ip address and performs any json configured UpdateIPAddress actions as required
func PerformDDNSActions(config *config.Configuration) error {
	var err error

	if config.Services == nil {
		log.Println("no DDNS services configured, nothing to do")
		return nil
	}

	ipAddrProvider := getPublicIpAddressProvider(&config.Router)
	currentPublicIpAddr, err = ipAddrProvider.GetPublicIPAddress()
	if err != nil {
		return err
	}

	if lastPublicIpAddr == nil || !currentPublicIpAddr.Equal(lastPublicIpAddr) {
		for _, serviceConfig := range config.Services {
			switch serviceConfig.ServiceType {
			case "DuckDNS":
				{
					err = updateIpAddress(currentPublicIpAddr, ddns.DuckDNSClient{ServiceConfig: serviceConfig})
					if err != nil {
						break
					}
				}
			case "Namecheap":
				{
					err = updateIpAddress(currentPublicIpAddr, ddns.NamecheapClient{ServiceConfig: serviceConfig})
					if err != nil {
						break
					}
				}
			case "NoIP":
				{
					err = updateIpAddress(currentPublicIpAddr, ddns.NoIPClient{ServiceConfig: serviceConfig})
					if err != nil {
						break
					}
				}
			}

		}
	} else {
		log.Printf("Public IPv4 address %s remains unchanged, no DDNS updates performed", currentPublicIpAddr)
	}

	if err == nil {
		lastPublicIpAddr = currentPublicIpAddr
	}

	return err
}

//returns a ipaddress.AddressProvider for the supplied routerConfig *config.RouterConfiguration
func getPublicIpAddressProvider(routerConfig *config.RouterConfiguration) ipaddress.AddressProvider {
	var ipAddressProvider ipaddress.AddressProvider
	if routerConfig != nil && routerConfig.RouterType != "" {
		switch routerConfig.RouterType {
		case "BTSmartHub2":
			ipAddressProvider = ipaddress.BTSmartHub2{Config: routerConfig}
		}
	} else {
		ipAddressProvider = ipaddress.Default{}
	}
	return ipAddressProvider
}

//performs an IPv4 ip address update using the supplied publicIpAddress and client
func updateIpAddress(publicIpAddress net.IP, client ddns.IDynamicDnsClient) error {
	return client.UpdateIPAddress(publicIpAddress)
}
