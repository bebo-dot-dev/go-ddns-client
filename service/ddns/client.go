package ddns

import (
	"go-ddns-client/service/config"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

//describes the interface of a type that knows how to perform a dynamic dns IP address update
type IDynamicDnsClient interface {
	//returns the name of a dynamic dns client
	Name() string
	//performs the dynamic dns IP address update operation
	UpdateIPAddress(publicIpAddr net.IP) error
	//logs the dynamic dns client IP address update
	LogIPAddressUpdate()
}

type Client struct {
	ServiceConfig config.ServiceConfiguration
}

//performs a HTTP GET request and returns the response
func PerformHttpRequest(url string, username string, password string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if username != "" && password != "" {
		request.SetBasicAuth(username, password)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return responseBytes, nil
}
