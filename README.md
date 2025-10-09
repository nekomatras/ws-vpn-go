# Notes

## Build and run

### Build:
- CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ws-vpn main.go

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

## TODO:

### Server:
  1. Вынести register страничку в отдельный модулю
  2. Вынести заглушку и добавить ее кастомизацию