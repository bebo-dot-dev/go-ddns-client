package ddns

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
)

// NamecheapClient implements the namecheap dynamic dns client
/*
namecheap docs: https://www.namecheap.com/support/knowledgebase/article.aspx/29/11/how-do-i-use-a-browser-to-dynamically-update-the-hosts-ip/

request url: https://dynamicdns.park-your-domain.com/update?host=@&domain=example.com&password=PASSWORD&ip=255.255.255.255

sample response xml:
<?xml version="1.0"?>
<interface-response>
    <Command>SETDNSHOST</Command>
    <Language>eng</Language>
    <ErrCount>1</ErrCount>
    <errors>
        <Err1>Error message here</Err1>
    </errors>
    <ResponseCount>1</ResponseCount>
    <responses>
        <response>
            <ResponseNumber>99999</ResponseNumber>
            <ResponseString>Error message here</ResponseString>
        </response>
    </responses>
    <Done>true</Done>
    <debug><![CDATA[]]></debug>
</interface-response>
*/
type NamecheapClient Client

type NamecheapXmlResponse struct {
	XMLName  xml.Name `xml:"interface-response"`
	Text     string   `xml:",chardata"`
	Command  string   `xml:"Command"`
	Language string   `xml:"Language"`
	ErrCount int      `xml:"ErrCount"`
	Errors   struct {
		Text string `xml:",chardata"`
		Err1 string `xml:"Err1"`
	} `xml:"errors"`
	ResponseCount string `xml:"ResponseCount"`
	Responses     struct {
		Text     string `xml:",chardata"`
		Response struct {
			Text           string `xml:",chardata"`
			ResponseNumber int    `xml:"ResponseNumber"`
			ResponseString string `xml:"ResponseString"`
		} `xml:"response"`
	} `xml:"responses"`
	Done  string `xml:"Done"`
	Debug string `xml:"debug"`
}

// Name returns the name of this dynamic dns client
func (client NamecheapClient) Name() string {
	return "Namecheap dynamic DNS client"
}

// UpdateIPAddress performs the dynamic dns IP address update operation
func (client NamecheapClient) UpdateIPAddress(publicIpAddress net.IP) error {
	dynDnsIpUpdateUrl := fmt.Sprintf(
		"https://dynamicdns.park-your-domain.com/update?host=@&domain=%s&password=%s&ip=%s",
		client.ServiceConfig.TargetDomain,
		client.ServiceConfig.Password,
		publicIpAddress)

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
	var namecheapXml NamecheapXmlResponse
	err = xml.Unmarshal(responseBytes, &namecheapXml)
	if err != nil {
		return err
	}
	if namecheapXml.ErrCount != 0 {
		err = errors.New(fmt.Sprintf("The namecheap IP address update to %s for domain %s failed with error: '%s', responseNumber: %d, responseString: '%s'",
			publicIpAddress, client.ServiceConfig.TargetDomain, namecheapXml.Errors.Err1, namecheapXml.Responses.Response.ResponseNumber, namecheapXml.Responses.Response.ResponseString))
		return err
	}

	client.LogIPAddressUpdate()

	return nil
}

// LogIPAddressUpdate logs the dynamic dns client IP address update
func (client NamecheapClient) LogIPAddressUpdate() {
	log.Printf("The %s IP address update for domain %s succeeded", client.Name(), client.ServiceConfig.TargetDomain)
}
