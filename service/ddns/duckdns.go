package ddns

import (
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

//String implements the Stringer interface to return the name of this dynamic dns client
func (client DuckDNSClient) String() string {
	return "DuckDNS dynamic DNS client"
}

// UpdateIPAddresses performs the dynamic dns IP address update operation
func (client DuckDNSClient) UpdateIPAddresses(ipv4, ipv6 net.IP) error {
	dynDnsIpUpdateUrl := fmt.Sprintf(
		"https://www.duckdns.org/update?domains=%s&token=%s&ip=%s&ipv6=%s",
		client.ServiceConfig.TargetDomain,
		client.ServiceConfig.Token,
		ipv4,
		ipv6)

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
		return fmt.Errorf("the DuckDNS IP address update to %s / %s for domain %s failed: '%s'",
			ipv4, ipv6, client.ServiceConfig.TargetDomain, responseStr)
	}

	client.LogIPAddressUpdate()

	return nil
}

// LogIPAddressUpdate logs the dynamic dns client IP address update
func (client DuckDNSClient) LogIPAddressUpdate() {
	log.Printf("The %s IP address update for domain %s succeeded", client, client.ServiceConfig.TargetDomain)
}
