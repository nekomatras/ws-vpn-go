# Notes

## Build and run

### Build:
- CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./target/ws-vpn main.go

### Cert:
- openssl req -x509 -newkey rsa:2048 -nodes -keyout ./target/key.pem -out ./target/cert.pem -days 365

### Apply env vars
- export $(grep -v '^#' client.env | xargs)

### Run:
- sudo ./ws-vpn -config=./client.conf

### Build Docker:
- sudo docker build --network=host --rm -t ws-vpn . && sudo docker image prune -f

### Export Image:
- sudo docker save -o ws-vpn.tar ws-vpn:latest

### Import Image:
- sudo docker load -i /tmp/ws-vpn.tar

### Run Image:
- sudo docker run --rm ws-vpn:latest
- sudo docker run --rm -v /path/to/config/config.conf:/etc/ws-wpn.conf app:latest

## TODO:

### Server:
  1. Вынести register страничку в отдельный модулю
  2. Вынести заглушку и добавить ее кастомизацию