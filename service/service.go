package service

import (
	"ddns-client/service/configuration"
	"ddns-client/service/ddnsClientProvider"
	"ddns-client/service/publicIPv4Provider"
	"log"
	"net"
)

var (
	currentPublicIpAddr net.IP
	lastPublicIpAddr    net.IP
)

//retrieves the current public IPv4 ip address and performs any json configured UpdateIPAddress actions as required
func PerformDDNSActions(config *configuration.Configuration) error {
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
					err = updateIpAddress(currentPublicIpAddr, ddnsClientProvider.DuckDNSClient{ServiceConfig: serviceConfig})
					if err != nil {
						break
					}
				}
			case "Namecheap":
				{
					err = updateIpAddress(currentPublicIpAddr, ddnsClientProvider.NamecheapClient{ServiceConfig: serviceConfig})
					if err != nil {
						break
					}
				}
			case "NoIP":
				{
					err = updateIpAddress(currentPublicIpAddr, ddnsClientProvider.NoIPClient{ServiceConfig: serviceConfig})
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

//returns a publicIPv4Provider.PublicIPAddressProvider for the supplied routerConfig *configuration.RouterConfiguration
func getPublicIpAddressProvider(routerConfig *configuration.RouterConfiguration) publicIPv4Provider.PublicIPAddressProvider {
	var ipAddressProvider publicIPv4Provider.PublicIPAddressProvider
	if routerConfig != nil && routerConfig.RouterType != "" {
		switch routerConfig.RouterType {
		case "BTSmartHub2":
			ipAddressProvider = publicIPv4Provider.BTSmartHub2IPAddressProvider{Config: routerConfig}
		}
	} else {
		ipAddressProvider = publicIPv4Provider.DefaultIPAddressProvider{}
	}
	return ipAddressProvider
}

//performs an IPv4 ip address update using the supplied publicIpAddress and client
func updateIpAddress(publicIpAddress net.IP, client ddnsClientProvider.IDynamicDnsClient) error {
	return client.UpdateIPAddress(publicIpAddress)
}
