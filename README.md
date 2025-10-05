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

### Client:
  ?1. Клиент должен получать MTU от сервера при открытии тунеля и устанавливать его в свой интерфейс после этого
  2. Если отваливается тунель WS, дропаем пакеты с интерфейса и пытаемся реконнектнуться

### Server:
  1. В хендлере в цикле читаем пакеты от клиента и сразу кидаем в интерфейс
  ?2. На этапе авторизации получайм от клиента ключ и его локальный ip (нужно проверять, что ip реально его...)
  3. Создаем канал из клиентского WS и кладем его в мапу по его ip
  4. В хендлере читаем из канала и кидаем все в канал интерфейса
  5. Создаем поток, который будет читать интерфейс и каждый входящий пакет будет кидать в соответствующий ip назначения WS канал