package ddns

import (
	"errors"
	"fmt"
	"go-ddns-client/service/notifications"
	"log"
	"net"
	"net/http"
	"strings"
)

// NoIPClient implements the noip dynamic dns client
/*
noip docs: https://www.noip.com/integrate/request

request url: https://username:password@dynupdate.no-ip.com/nic/update?hostname=mytest.example.com&myip=192.0.2.25

sample response:
nochg 192.0.2.25
*/
type NoIPClient Client

// Name returns the name of this dynamic dns client
func (client NoIPClient) Name() string {
	return "NoIP dynamic DNS client"
}

// UpdateIPAddress performs the dynamic dns IP address update operation
func (client NoIPClient) UpdateIPAddress(publicIpAddress net.IP) error {
	dynDnsIpUpdateUrl := fmt.Sprintf(
		"https://dynupdate.no-ip.com/nic/update?hostname=%s&myip=%s",
		client.ServiceConfig.TargetDomain,
		publicIpAddress)

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
		return errors.New(fmt.Sprintf("The noIP IP address update to %s for domain %s failed: '%s'",
			publicIpAddress, client.ServiceConfig.TargetDomain, responseStr))
	}

	notifications.GetManager(client.NotificationConfig).Send(client.ServiceConfig.TargetDomain, publicIpAddress.String())
	client.LogIPAddressUpdate()

	return nil
}

// LogIPAddressUpdate logs the dynamic dns client IP address update
func (client NoIPClient) LogIPAddressUpdate() {
	log.Printf("The %s IP address update for domain %s succeeded", client.Name(), client.ServiceConfig.TargetDomain)
}
