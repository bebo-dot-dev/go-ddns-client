package ddns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

// CloudFlareClient implements the cloudflare dynamic dns client
/*
cloudflare docs:
	https://api.cloudflare.com/#getting-started-resource-ids
	https://api.cloudflare.com/#dns-records-for-a-zone-list-dns-records
	https://api.cloudflare.com/#dns-records-for-a-zone-update-dns-record
*/
type CloudFlareClient Client

type ZonesJsonResponse struct {
	Zones []struct {
		ID string `json:"id"`
	} `json:"result"`
	Success bool `json:"success"`
}

type ListDnsRecordsResponse struct {
	Success    bool        `json:"success"`
	DnsRecords []DnsRecord `json:"result"`
}

type DnsRecord struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	ZoneID   string `json:"zone_id"`
	ZoneName string `json:"zone_name"`
}

type DnsRecordUpdateResponse struct {
	Success bool `json:"success"`
}

// UpdateIPAddresses performs the dynamic dns IP address update operation
func (client CloudFlareClient) UpdateIPAddresses(ipv4, ipv6 net.IP) error {
	zonesResp, err := client.getZones()
	if err != nil {
		return err
	}

	dnsRecordsUpdated := 0
	for _, zone := range zonesResp.Zones {
		dnsRecordsResp, err := client.getDnsRecords(zone.ID)
		if err != nil {
			return err
		}
		for _, dnsRecord := range dnsRecordsResp.DnsRecords {
			err := client.applyIpAddressUpdate(&dnsRecord, ipv4, ipv6)
			if err != nil {
				return err
			}
			dnsRecordsUpdated++
		}
	}

	Client(client).LogIPAddressUpdate(fmt.Sprintf("%d DNS records were updated", dnsRecordsUpdated))
	log.Printf("%d DNS records were updated at %s", dnsRecordsUpdated, client.ServiceConfig.ServiceType)

	return nil
}

//getRequestHeaders returns headers commonly used across all Cloudflare requests
func (client CloudFlareClient) getRequestHeaders() *map[string]string {
	headers := make(map[string]string)
	headers["X-Auth-Email"] = client.ServiceConfig.EmailAddress
	headers["X-Auth-Key"] = client.ServiceConfig.APIKey
	headers["accept"] = "application/json"
	headers["Content-Type"] = "application/json"
	return &headers
}

//getZones returns a *ZonesJsonResponse representing all configured zones at Cloudflare
//for the specified client.ServiceConfig.EmailAddress / client.ServiceConfig.APIKey
func (client CloudFlareClient) getZones() (*ZonesJsonResponse, error) {
	headers := client.getRequestHeaders()

	statusCode, responseBytes, err := PerformHttpRequest(
		http.MethodGet,
		"https://api.cloudflare.com/client/v4/zones",
		"",
		"",
		nil,
		*headers)

	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("%s zones GET returned http status code %d and response \n%s",
			client.ServiceConfig.ServiceType, statusCode, string(responseBytes))
	}

	var zonesJson ZonesJsonResponse
	if err = json.Unmarshal(responseBytes, &zonesJson); err != nil {
		return nil, err
	}

	if !zonesJson.Success {
		return nil, fmt.Errorf("%s zones GET returned an unsuccessful response \n%s",
			client.ServiceConfig.ServiceType, string(responseBytes))
	}

	log.Printf("%s zones GET returned %d zone(s) for processing",
		client.ServiceConfig.ServiceType, len(zonesJson.Zones))

	return &zonesJson, nil
}

//getDnsRecords returns a *ListDnsRecordsResponse representing all DNS records within the specified zoneId
func (client CloudFlareClient) getDnsRecords(zoneId string) (*ListDnsRecordsResponse, error) {
	headers := client.getRequestHeaders()

	statusCode, responseBytes, err := PerformHttpRequest(
		http.MethodGet,
		fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zoneId),
		"",
		"",
		nil,
		*headers)

	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("%s dns_records GET returned http status code %d and response \n%s",
			client.ServiceConfig.ServiceType, statusCode, string(responseBytes))
	}

	var dnsRecordsJson ListDnsRecordsResponse
	if err = json.Unmarshal(responseBytes, &dnsRecordsJson); err != nil {
		return nil, err
	}

	if !dnsRecordsJson.Success {
		return nil, fmt.Errorf("%s dns_records GET returned an unsuccessful response \n%s",
			client.ServiceConfig.ServiceType, string(responseBytes))
	}

	log.Printf("%s dns_records GET returned %d dns record(s) for processing",
		client.ServiceConfig.ServiceType, len(dnsRecordsJson.DnsRecords))

	return &dnsRecordsJson, nil
}

//applyIpAddressUpdate applies the DNS record update
func (client CloudFlareClient) applyIpAddressUpdate(dnsRecord *DnsRecord, ipv4, ipv6 net.IP) error {
	var err error
	if strings.EqualFold(client.ServiceConfig.TargetDomain, dnsRecord.Name) ||
		strings.EqualFold(client.ServiceConfig.TargetDomain, dnsRecord.ZoneName) {
		var ipToUse *net.IP
		switch dnsRecord.Type {
		case "A":
			ipToUse = &ipv4
		case "AAAA":
			ipToUse = &ipv6
		}
		if ipToUse != nil {
			headers := client.getRequestHeaders()

			jsonBody := fmt.Sprintf(`{
				"type":"%s",
				"name":"%s",
				"content":"%s",
				"ttl":%d
			}`, dnsRecord.Type, client.ServiceConfig.TargetDomain, *ipToUse, client.ServiceConfig.TTL)

			statusCode, responseBytes, err := PerformHttpRequest(
				http.MethodPut,
				fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", dnsRecord.ZoneID, dnsRecord.ID),
				"",
				"",
				bytes.NewBuffer([]byte(jsonBody)),
				*headers)

			if err != nil {
				return err
			}

			if statusCode != http.StatusOK {
				return fmt.Errorf("%s dns_records PUT returned http status code %d and response \n%s",
					client.ServiceConfig.ServiceType, statusCode, string(responseBytes))
			}

			var dnsUpdateJson DnsRecordUpdateResponse
			if err = json.Unmarshal(responseBytes, &dnsUpdateJson); err != nil {
				return err
			}

			if !dnsUpdateJson.Success {
				return fmt.Errorf("%s dns_records PUT returned an unsuccessful response \n%s",
					client.ServiceConfig.ServiceType, string(responseBytes))
			}
		}
	}
	return err
}
