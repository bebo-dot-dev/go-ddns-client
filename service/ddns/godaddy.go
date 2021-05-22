package ddns

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
)

/*
The godaddy dynamic dns client

godaddy docs: https://developer.godaddy.com/doc/endpoint/domains#/v1/recordReplaceTypeName

request urls:
	https://api.ote-godaddy.com/v1/domains/example.com/records/A/recordName
	https://api.godaddy.com/v1/domains/example.com/records/A/recordName

curl:
	curl -X PUT "https://api.ote-godaddy.com/v1/domains/example.com/records/A/recordName" -H  "accept: application/json" -H  "Content-Type: application/json" -H  "Authorization: sso-key UzQxLikm_46KxDFnbjN7cQjmw6wocia:46L26ydpkwMaKZV6uVdDWe" -d "[  {    \"data\": \"127.0.0.1\",    \"port\": 53,    \"priority\": 0,    \"protocol\": \"string\",    \"service\": \"string\",    \"ttl\": 600,    \"weight\": 0  }]"

sample response:
{
  "code": "ACCESS_DENIED",
  "message": "Authenticated user is not allowed access"
}
*/
type GoDaddyClient Client

//returns the name of this dynamic dns client
func (client GoDaddyClient) Name() string {
	return "GoDaddy API dynamic DNS client"
}

//performs the dynamic dns IP address update operation
func (client GoDaddyClient) UpdateIPAddress(publicIpAddress net.IP) error {
	dynDnsIpUpdateUrl := fmt.Sprintf(
		"https://api.godaddy.com/v1/domains/%s/records/A/%s",
		client.ServiceConfig.TargetDomain,
		client.ServiceConfig.RecordName)

	jsonBody := fmt.Sprintf(`[{
		"data": "%s",
		"port": %d,
		"priority": 0,
		"protocol": "string",
		"service": "string",
		"ttl": %d,
		"weight": 0
	  }]`, publicIpAddress, client.ServiceConfig.Port, client.ServiceConfig.TTL)
	
	headers := make(map[string]string)
	headers["accept"] = "application/json"
	headers["Content-Type"] = "application/json"
	headers["Authorization"] = fmt.Sprintf("sso-key %s:%s", client.ServiceConfig.APIKey, client.ServiceConfig.APISecret)

	statusCode, responseBytes, err := PerformHttpRequest(
		http.MethodPut,
		dynDnsIpUpdateUrl,
		"",
		"",
		bytes.NewBuffer([]byte(jsonBody)),
		headers)

	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		responseStr := string(responseBytes)
		return errors.New(fmt.Sprintf("The GoDaddy IP address update to %s for domain %s failed: '%s'",
			publicIpAddress, client.ServiceConfig.TargetDomain, responseStr))
	}

	client.LogIPAddressUpdate()
	return nil
}

//logs the dynamic dns client IP address update
func (client GoDaddyClient) LogIPAddressUpdate() {
	log.Printf("The %s IP address update for domain %s succeeded", client.Name(), client.ServiceConfig.TargetDomain)
}
