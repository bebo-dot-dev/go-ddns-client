# go-ddns-client
A Go dynamic DNS client
------
### Features:
* Application behaviour, router, DDNS service and notification service configuration implemented [in json](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/serviceConfig.json)
* Configuration change detection, hot refresh / reload  
* Public IP address determination via direct communication with a 
  [BT Smart Hub 2 router](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/service/ipaddress/btsmarthub2.go) on the LAN, prevents having to perform an HTTP request to a public external internet service to determine the current public IP.
* [Falls back](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/service/ipaddress/default.go) to using https://api.ipify.org to determine the public IP address when not using a BT Smart Hub 2 router.
* Realtime notifications
### Supported DDNS services:
* [DuckDNS](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/service/ddns/duckdns.go)
* [GoDaddy](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/service/ddns/godaddy.go)
* [Namecheap](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/service/ddns/namecheap.go)
* [NoIP](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/service/ddns/noip.go)
### Supported notification services:
* [Email (SSL and TLS)](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/service/notifications/email.go)
* [Sipgate IO SMS](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/service/notifications/sipgate.go)
### Tested on:
* Linux x64
* Linux Arm aarch64
* Windows10
### Build with Go build:
```bash
#Linux x64
env GOOS=linux go build ./cmd/go-ddns-client/go-ddns-client.go
```
```bash
#Linux Arm aarch64
env GOOS=linux GOARCH=arm64 go build ./cmd/go-ddns-client/go-ddns-client.go
```
```bash
#Windows
env GOOS=windows go build ./cmd/go-ddns-client/go-ddns-client.go
```