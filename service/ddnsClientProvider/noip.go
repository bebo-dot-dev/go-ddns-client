package ddnsClientProvider

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

/*
The noip dynamic dns client

noip docs: https://www.noip.com/integrate/request

request url: https://username:password@dynupdate.no-ip.com/nic/update?hostname=mytest.example.com&myip=192.0.2.25

sample response:
nochg 192.0.2.25
*/
type NoIPClient DynamicDnsClient

//returns the name of this dynamic dns client
func (client NoIPClient) Name() string {
	return "NoIP dynamic DNS client"
}

//performs the dynamic dns IP address update operation
func (client NoIPClient) UpdateIPAddress(publicIpAddress net.IP) error {
	dynDnsIpUpdateUrl := fmt.Sprintf(
		"https://dynupdate.no-ip.com/nic/update?hostname=%s&myip=%s",
		client.ServiceConfig.TargetDomain,
		publicIpAddress)

	responseBytes, err := PerformHttpRequest(
		dynDnsIpUpdateUrl,
		client.ServiceConfig.Username,
		client.ServiceConfig.Password)

	if err != nil {
		fmt.Println(responseBytes)
		return err
	}

	responseStr := string(responseBytes)
	if !strings.HasPrefix(responseStr, "nochg") && !strings.HasPrefix(responseStr, "good") {
		return errors.New(fmt.Sprintf("The noIP IP address update to %s for domain %s failed: '%s'",
			publicIpAddress, client.ServiceConfig.TargetDomain, responseStr))
	}

	client.LogIPAddressUpdate()
	return nil
}

//logs the dynamic dns client IP address update
func (client NoIPClient) LogIPAddressUpdate() {
	log.Printf("The %s IP address update for domain %s succeeded", client.Name(), client.ServiceConfig.TargetDomain)
}
