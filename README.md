# go-ddns-client
A Go dynamic DNS client
------
### Features:
* Application behaviour, router and DDNS service configuration implemented in json
* Implements a method to determine the public IP address currently in use by direct communication with a 
  BT Smart Hub 2 router on the LAN. This feature removes the need to perform an HTTP request to a public external 
  internet service to determine the current public IP.
* Falls back to using https://api.ipify.org to determine the public IP address when not using a BT Smart Hub 2 router.
### Supported DDNS services:
* DuckDNS
* Namecheap
* NoIP
### Tested on:
* Linux x64
* Linux Arm aarch64
### Build with Go build:
```bash
#Linux x64
env GOOS=linux go build ./cmd/ddns-client/ddns-client.go
```
```bash
#Linux Arm aarch64
env GOOS=linux GOARCH=arm64 go build ./cmd/ddns-client/ddns-client.go
```


