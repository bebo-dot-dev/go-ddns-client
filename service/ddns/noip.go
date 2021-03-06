package ddns

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// NoIPClient implements the noip dynamic dns client
/*
noip docs: https://www.noip.com/integrate/request

request url: https://username:password@dynupdate.no-ip.com/nic/update?hostname=mytest.example.com&myip=192.0.2.25&myipv6=2a0f:21a1:2103:2001:f5e:1111:6fd:6bc7

sample response:
nochg 192.0.2.25
*/
type NoIPClient Client

// UpdateIPAddresses performs the dynamic dns IP address update operation
func (client NoIPClient) UpdateIPAddresses(ipv4, ipv6 net.IP) error {
	dynDnsIpUpdateUrl := fmt.Sprintf(
		"https://dynupdate.no-ip.com/nic/update?hostname=%s&myip=%s&myipv6=%s",
		client.ServiceConfig.TargetDomain,
		ipv4,
		ipv6)

	_, responseBytes, err := PerformHttpRequest(
		http.MethodGet,
		dynDnsIpUpdateUrl,
		client.ServiceConfig.Username,
		client.ServiceConfig.Password,
		nil,
		nil)

	if err != nil {
		fmt.Println(responseBytes)
		return err
	}

	responseStr := string(responseBytes)
	if !strings.HasPrefix(responseStr, "nochg") && !strings.HasPrefix(responseStr, "good") {
		return fmt.Errorf("the noIP IP address update to %s / %s for domain %s failed: '%s'",
			ipv4, ipv6, client.ServiceConfig.TargetDomain, responseStr)
	}

	Client(client).LogIPAddressUpdate()

	return nil
}
