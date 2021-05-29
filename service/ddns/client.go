package ddns

import (
	"go-ddns-client/service/config"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

// IDynamicDnsClient describes the interface of a type that knows how to perform a dynamic dns IP address update
type IDynamicDnsClient interface {
	// UpdateIPAddress performs the dynamic dns IP address update operation
	UpdateIPAddress(publicIpAddr net.IP) error
	// LogIPAddressUpdate logs the dynamic dns client IP address update
	LogIPAddressUpdate()
}

type Client struct {
	ServiceConfig *config.ServiceConfiguration
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

	if headers != nil {
		for key, value := range headers {
			request.Header.Set(key, value)
		}
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
