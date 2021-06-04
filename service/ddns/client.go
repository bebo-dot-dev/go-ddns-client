package ddns

import (
	"github.com/bebo-dot-dev/go-ddns-client/service/config"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// IDynamicDnsClient describes the interface of a type that knows how to perform a dynamic dns IP address update
type IDynamicDnsClient interface {
	// UpdateIPAddresses performs the dynamic dns IP address update operation
	UpdateIPAddresses(ipv4, ipv6 net.IP) error
}

type Client struct {
	ServiceConfig *config.ServiceConfiguration
}

// LogIPAddressUpdate logs the dynamic dns client IP address update
func (client Client) LogIPAddressUpdate(args ...string) {
	log.Printf("The %s IP address update for domain %s succeeded. %v",
		client.ServiceConfig.ServiceType, client.ServiceConfig.TargetDomain, strings.Join(args, ","))
}

// PerformHttpRequest performs a HTTP request and returns the status code and the response
func PerformHttpRequest(
	method string,
	url string,
	username string,
	password string,
	body io.Reader,
	headers map[string]string) (int, []byte, error) {

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return 0, nil, err
	}

	if username != "" && password != "" {
		request.SetBasicAuth(username, password)
	}

	for key, value := range headers {
		request.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return 0, nil, err
	}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, nil, err
	}

	return response.StatusCode, responseBytes, nil
}
