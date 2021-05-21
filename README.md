# go-ddns-client
A Go dynamic DNS client
------
### Features:
* Application behaviour, router and DDNS service configuration implemented [in json](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/serviceConfig.json)
* Implements [a method](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/service/ipaddress/btsmarthub2.go) to determine the public IP address currently in use by direct communication with a 
  [BT Smart Hub 2 router](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/service/ipaddress/btsmarthub2.go) on the LAN. This feature removes the need to perform an HTTP request to a public external 
  internet service to determine the current public IP.
* [Falls back](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/service/ipaddress/default.go) to using https://api.ipify.org to determine the public IP address when not using a BT Smart Hub 2 router.
### Supported DDNS services:
* [DuckDNS](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/service/ddns/duckdns.go)
* [Namecheap](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/service/ddns/namecheap.go)
* [NoIP](https://github.com/bebo-dot-dev/go-ddns-client/blob/main/service/ddns/noip.go)
### Tested on:
* Linux x64
* Linux Arm aarch64
### Build with Go build:
```bash
#Linux x64
env GOOS=linux go build ./cmd/go-ddns-client/go-ddns-client.go
```
```bash
#Linux Arm aarch64
env GOOS=linux GOARCH=arm64 go build ./cmd/go-ddns-client/go-ddns-client.go
```


