Build:
- CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ws-vpn main.go

Run:
- sudo ./ws-vpn -mode=client -remote=ws://8.8.8.8:8080/ws

Build Docker:
- sudo docker build --network=host --rm -t ws-vpn . && sudo docker image prune -f

Export Image:
- sudo docker save -o ws-vpn.tar ws-vpn:latest

Import Image:
- sudo docker load -i /tmp/ws-vpn.tar

Run Image:
- sudo docker run --rm ws-vpn:latest