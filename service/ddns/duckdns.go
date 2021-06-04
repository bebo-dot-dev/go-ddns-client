package ddns

import (
	"fmt"
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

	Client(client).LogIPAddressUpdate()

	return nil
}
