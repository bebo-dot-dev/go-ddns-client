package ddns

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
)

// DuckDNSClient implements the duckdns dynamic dns client
/*
duckdns docs: https://www.duckdns.org/spec.jsp

request url: https://www.duckdns.org/update?domains={YOURVALUE}&token={YOURVALUE}[&ip={YOURVALUE}][&ipv6={YOURVALUE}][&verbose=true][&clear=true]

sample response:
OK
*/
type DuckDNSClient Client

// Name returns the name of this dynamic dns client
func (client DuckDNSClient) Name() string {
	return "DuckDNS dynamic DNS client"
}

// UpdateIPAddress performs the dynamic dns IP address update operation
func (client DuckDNSClient) UpdateIPAddress(publicIpAddress net.IP) error {
	dynDnsIpUpdateUrl := fmt.Sprintf(
		"https://www.duckdns.org/update?domains=%s&token=%s&ip=%s",
		client.ServiceConfig.TargetDomain,
		client.ServiceConfig.Token,
		publicIpAddress)

	_, responseBytes, err := PerformHttpRequest(
		http.MethodGet,
		dynDnsIpUpdateUrl,
		"",
		"",
		nil,
		nil)

	if err != nil {
		fmt.Println(responseBytes)
		return err
	}

	responseStr := string(responseBytes)
	if responseStr != "OK" {
		return errors.New(fmt.Sprintf("The DuckDNS IP address update to %s for domain %s failed: '%s'",
			publicIpAddress, client.ServiceConfig.TargetDomain, responseStr))
	}

	client.LogIPAddressUpdate()

	return nil
}

// LogIPAddressUpdate logs the dynamic dns client IP address update
func (client DuckDNSClient) LogIPAddressUpdate() {
	log.Printf("The %s IP address update for domain %s succeeded", client.Name(), client.ServiceConfig.TargetDomain)
}
