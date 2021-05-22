package ipaddress

import (
	"io"
	"log"
	"net"
	"net/http"
)

// AddressProvider describes the interface of a type able to return the public facing IP address in use where this code is running
type AddressProvider interface {
	// ProviderName returns the name of an IPv4 public IP address provider
	ProviderName() string
	// GetPublicIPAddress returns the public IP address
	GetPublicIPAddress() (net.IP, error)
	// LogPublicIPAddress logs the public IP address
	LogPublicIPAddress(net.IP)
}

/*
The Default type that has the ability to talk to api.ipify.org to retrieve a public IPv4 address
This type acts as a fallback in the event of there being no configured json router section publicIPv4Provider

request url: https://api.ipify.org

sample json response: 255.255.255.255
*/
type Default struct{}

// ProviderName returns the name of this IPv4 public IP address provider
func (ipProvider Default) ProviderName() string {
	return "api.ipify.org public IPV4 address provider"
}

// GetPublicIPAddress performs a HTTP request to https://api.ipify.org to retrieve and return the public IP address
func (ipProvider Default) GetPublicIPAddress() (net.IP, error) {
	response, err := http.Get("https://api.ipify.org")
	if err != nil {
		return nil, err
	}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	ipBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var ipv4 = net.IP{}
	err = ipv4.UnmarshalText(ipBytes)
	if err != nil {
		return nil, err
	}
	ipProvider.LogPublicIPAddress(ipv4)
	return ipv4, nil
}

// LogPublicIPAddress logs the public IP address
func (ipProvider Default) LogPublicIPAddress(ip net.IP) {
	log.Printf("The %s reports the public IPv4 as %s", ipProvider.ProviderName(), ip)
}
