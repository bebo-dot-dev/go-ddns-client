package ddns

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
)

// GoDaddyClient implements the godaddy dynamic dns client
/*
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

// UpdateIPAddresses performs the dynamic dns IP address update operation
func (client GoDaddyClient) UpdateIPAddresses(ipv4, ipv6 net.IP) error {
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
	  }]`, ipv4, client.ServiceConfig.Port, client.ServiceConfig.TTL)

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
		return fmt.Errorf("the GoDaddy IP address update to %s for domain %s failed: '%s'",
			ipv4, client.ServiceConfig.TargetDomain, responseStr)
	}

	Client(client).LogIPAddressUpdate()

	return nil
}
