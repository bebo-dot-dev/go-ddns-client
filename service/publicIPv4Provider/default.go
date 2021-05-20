package publicIPv4Provider

import (
	"io"
	"log"
	"net"
	"net/http"
)

//Describes the interface of a type able to return the public facing IP address in use where this code is running
type PublicIPAddressProvider interface {
	//returns the name of an IPv4 public IP address provider
	ProviderName() string
	//returns the public IP address
	GetPublicIPAddress() (net.IP, error)
	//logs the public IP address
	LogPublicIPAddress(net.IP)
}

/*
The DefaultIPAddressProvider type that has the ability to talk to api.ipify.org to retrieve a public IPv4 address
This type acts as a fallback in the event of there being no configured json router section publicIPv4Provider

request url: https://api.ipify.org

sample json response: 255.255.255.255
*/
type DefaultIPAddressProvider struct{}

//returns the name of this IPv4 public IP address provider
func (ipProvider DefaultIPAddressProvider) ProviderName() string {
	return "api.ipify.org public IPV4 address provider"
}

//performs a HTTP request to https://api.ipify.org to retrieve and return the public IP address
func (ipProvider DefaultIPAddressProvider) GetPublicIPAddress() (net.IP, error) {
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

//logs the public IP address
func (ipProvider DefaultIPAddressProvider) LogPublicIPAddress(ip net.IP) {
	log.Printf("The %s reports the public IPv4 as %s", ipProvider.ProviderName(), ip)
}
