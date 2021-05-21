package ipaddress

import (
	"encoding/xml"
	"errors"
	"fmt"
	"go-ddns-client/service/config"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

/*
The BTSmartHub2 type that has the ability to talk to a BT Smart Hub 2 to retrieve a public IPv4 address

request url: http://192.168.1.254/nonAuth/wan_conn.xml

sample xml response:
<status>
    <!-- REAL -->
    <wan_conn_status_list type="array" value="[['connected%3B64%3Bpass'],
['disconnected%3B0%3Bpass'],
['disconnected%3B0%3Bpass'],
null]" />
    <wan_conn_volume_list type="array" value="[['297625188703%3B282899515695%3B14725673008'],
['0%3B0%3B0'],
['0%3B0%3B0'],
null]" />
    <wan_linestatus_rate_list type="array" value="[['DOWN','ADSL','G%2EDMT','0','0','0','0','0','0','0','0','0','0','0','0','fast'],
null]" />
    <wlan_channel_list type="array" value="[['6','0','6'],
['6','0','6'],
['6','0','6'],
['6','0','6'],
['6','0','6'],
['36','0','36'],
['36','0','36'],
['36','0','36'],
['36','0','36'],
['36','0','36'],
null]" />
    <curlinkstatus type="array" value="[['connected%3B47'],
['disconnected%3B0'],
['disconnected%3B0'],
null]" />
    <sysuptime value="849301" />
    <status_rate type="array" value="[['1000000000%3B1000000000%3B0%3B0'],
['0%3B0%3B0%3B0'],
['0%3B0%3B0%3B0'],
null]" />
    <wan_active_idx value="0" />
    <link_status value="disconnected%3Badsl%3B0" />
    <ip4_info_list type="array" value="[['81%2E255%2E255%2E255%3B255%2E255%2E255%2E255%3B172%2E16%2E13%2E79%3B81%2E139%2E56%2E100%3B81%2E139%2E57%2E100'],
['0%2E0%2E0%2E0%3B0%2E0%2E0%2E0%3B0%2E0%2E0%2E0%3B0%2E0%2E0%2E0%3B0%2E0%2E0%2E0'],
['0%2E0%2E0%2E0%3B0%2E0%2E0%2E0%3B0%2E0%2E0%2E0%3B0%2E0%2E0%2E0%3B0%2E0%2E0%2E0'],
null]" />
    <ip6_lla_list type="array" value="[['fe80%3A%3Afe80%3Afe80%3Afe80%3Afe80%2F10'],
['%3A%3A%2F0'],
['%3A%3A%2F0'],
null]" />
    <ip6_gua_list type="array" value="[['2a00%3A2a00%3A2a00%3A2a00%3A%3A1%2F64%3B2a00%3A%3A221%3A5ff%3A2a00%3A2a00'],
['%3A%3A%2F0%3B%3A%3A'],
['%3A%3A%2F0%3B%3A%3A'],
null]" />
    <ip6_rdns_list type="array" value="[[null],[null],[null]]" />
    <!-- END_REAL -->
    <!--for home page, login lock-->
    <locktime value="1" />
    <!--END for home page, login lock-->
</status>
*/
type BTSmartHub2 struct {
	Config *config.RouterConfiguration
}

//models the xml response returned by a BT smart hub 2 /nonAuth/wan_conn.xml request
type RouterStatus struct {
	XMLName           xml.Name `xml:"status"`
	Text              string   `xml:",chardata"`
	WanConnStatusList struct {
		Text  string `xml:",chardata"`
		Type  string `xml:"type,attr"`
		Value string `xml:"value,attr"`
	} `xml:"wan_conn_status_list"`
	WanConnVolumeList struct {
		Text  string `xml:",chardata"`
		Type  string `xml:"type,attr"`
		Value string `xml:"value,attr"`
	} `xml:"wan_conn_volume_list"`
	WanLinestatusRateList struct {
		Text  string `xml:",chardata"`
		Type  string `xml:"type,attr"`
		Value string `xml:"value,attr"`
	} `xml:"wan_linestatus_rate_list"`
	WlanChannelList struct {
		Text  string `xml:",chardata"`
		Type  string `xml:"type,attr"`
		Value string `xml:"value,attr"`
	} `xml:"wlan_channel_list"`
	Curlinkstatus struct {
		Text  string `xml:",chardata"`
		Type  string `xml:"type,attr"`
		Value string `xml:"value,attr"`
	} `xml:"curlinkstatus"`
	Sysuptime struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"sysuptime"`
	StatusRate struct {
		Text  string `xml:",chardata"`
		Type  string `xml:"type,attr"`
		Value string `xml:"value,attr"`
	} `xml:"status_rate"`
	WanActiveIdx struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"wan_active_idx"`
	LinkStatus struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"link_status"`
	Ip4InfoList struct {
		Text  string `xml:",chardata"`
		Type  string `xml:"type,attr"`
		Value string `xml:"value,attr"`
	} `xml:"ip4_info_list"`
	Ip6LlaList struct {
		Text  string `xml:",chardata"`
		Type  string `xml:"type,attr"`
		Value string `xml:"value,attr"`
	} `xml:"ip6_lla_list"`
	Ip6GuaList struct {
		Text  string `xml:",chardata"`
		Type  string `xml:"type,attr"`
		Value string `xml:"value,attr"`
	} `xml:"ip6_gua_list"`
	Ip6RdnsList struct {
		Text  string `xml:",chardata"`
		Type  string `xml:"type,attr"`
		Value string `xml:"value,attr"`
	} `xml:"ip6_rdns_list"`
	Locktime struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"locktime"`
}

//returns the name of this IPv4 public IP address provider
func (ipProvider BTSmartHub2) ProviderName() string {
	return "BTSmartHub2 public IPV4 address provider"
}

//performs a HTTP request to a BT smart hub 2 router to retrieve and return the public IP address
func (ipProvider BTSmartHub2) GetPublicIPAddress() (net.IP, error) {
	if ipProvider.Config == nil {
		return nil, errors.New("config is nil and it needs to be supplied")
	}
	xmlBytes, err := getRouterStatusXml(ipProvider.Config.IpDetailsUrl)
	if err != nil {
		return nil, err
	}

	var routerStatus RouterStatus
	err = xml.Unmarshal(xmlBytes, &routerStatus)
	if err != nil {
		return nil, err
	}
	ipv4sArr := strings.Split(routerStatus.Ip4InfoList.Value, ",")
	decodedIpv4s, err := url.QueryUnescape(strings.Trim(ipv4sArr[0], "[]'"))
	if err != nil {
		return nil, err
	}
	ipv4sArr = strings.Split(decodedIpv4s, ";")
	ipv4 := net.ParseIP(ipv4sArr[0])
	if ipv4 == nil {
		return nil, errors.New(fmt.Sprintf("unable to determine public ip from %s", decodedIpv4s))
	}
	ipProvider.LogPublicIPAddress(ipv4)
	return ipv4, nil
}

//logs the public IP address
func (ipProvider BTSmartHub2) LogPublicIPAddress(ip net.IP) {
	log.Printf("The %s reports the public IPv4 as %s", ipProvider.ProviderName(), ip)
}

//performs a HTTP GET request to retrieve and return the /nonAuth/wan_conn.xml
func getRouterStatusXml(ipDetailsUrl string) ([]byte, error) {
	response, err := http.Get(ipDetailsUrl)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	xmlBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return xmlBytes, nil
}
