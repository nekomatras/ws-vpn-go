# Notes

## Build and run

### Build:
- `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./target/ws-vpn main.go`

### Cert:
- `openssl req -x509 -newkey rsa:2048 -nodes -keyout ./target/key.pem -out ./target/cert.pem -days 365`

### Apply env vars
- `export $(grep -v '^#' client.env | xargs)`

### Run:
- `sudo ./ws-vpn -config -config-path=./client.conf`
Or:
- `sudo ./ws-vpn -env`

### Setup GeoIp:
Download geo ip data:
`curl -O https://www.ipdeny.com/ipblocks/data/countries/ru.zone`
Setup ipset:
`sudo pacman -S ipset`
Check that nftables is not running:
```
sudo systemctl status --now nftables
```
Create ipset:
`sudo ipset create ru hash:net`
Fill it:
```
for i in $(cat ru.zone); do
    sudo ipset add ru $i
done
```
And another one for local networks:
```
sudo ipset create local hash:net
sudo ipset add local 127.0.0.0/8
sudo ipset add local 10.0.0.0/8
sudo ipset add local 172.16.0.0/12
sudo ipset add local 192.168.0.0/16
sudo ipset add local 169.254.0.0/16
```
Check created set: `sudo ipset list local`

Create packet marking rule. Mark trafic by 1 if destination address is not in our ipsets:
```
sudo iptables -t mangle -A OUTPUT \
    -m set ! --match-set ru dst \
    -m set ! --match-set local dst \
    -j MARK --set-mark 1
```
Declare routing table in `/etc/iproute2/rt_tables`:
```
200 vpn
```

Set in config:
```
"route_table": 200
"redirect_by_mark": 1
```