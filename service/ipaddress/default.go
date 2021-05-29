package ipaddress

import (
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

// IAddressProvider describes the interface of a type able to return the public facing IP address in use where this code is running
type IAddressProvider interface {
	// GetPublicIPAddresses returns public IP addresses
	GetPublicIPAddresses() (net.IP, net.IP, error)
	// LogIPAddresses logs the public IP addresses
	LogIPAddresses(net.IP, net.IP)
}

/*
The Default type that has the ability to talk to api.ipify.org to retrieve a public IPv4 address
This type acts as a fallback in the event of there being no configured json router section publicIPv4Provider

request url: https://api.ipify.org

sample json response: 255.255.255.255
*/
type Default struct{}

//String implements the Stringer interface to return the name of this IAddressProvider
func (ipProvider Default) String() string {
	return "api.ipify.org IP address provider"
}

// GetPublicIPAddresses performs a HTTP request to https://api.ipify.org to retrieve and return the public IPv4 address
// and calls GetIPv6 to return the current IPv6 address of the host where this code is executing
func (ipProvider Default) GetPublicIPAddresses() (net.IP, net.IP, error) {
	response, err := http.Get("https://api.ipify.org")
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	ipBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, err
	}
	var ipv4 = net.IP{}
	err = ipv4.UnmarshalText(ipBytes)
	if err != nil {
		return nil, nil, err
	}

	ipv6, err := GetIPv6()
	if err != nil {
		return nil, nil, err
	}

	ipProvider.LogIPAddresses(ipv4, ipv6)
	return ipv4, ipv6, nil
}

//GetIPv6 returns the IPv6 address of the host where this code is executing
func GetIPv6() (net.IP, error) {
	var ip net.IP
	ifAddresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, address := range ifAddresses {
		if ipNetwork, ok := address.(*net.IPNet); ok && !ipNetwork.IP.IsLoopback() {
			if len(strings.Split(ipNetwork.IP.String(), ":")) == 8 {
				ip = ipNetwork.IP
				break
			}
		}
	}
	return ip, nil
}

// LogIPAddresses logs the public IP addresses
func (ipProvider Default) LogIPAddresses(ipv4, ipv6 net.IP) {
	log.Printf("The %s reports the public IPv4 as %s and the public IPv6 as %s", ipProvider, ipv4, ipv6)
}
