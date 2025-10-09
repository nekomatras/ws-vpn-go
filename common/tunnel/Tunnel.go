package tunnel

import (
	"io"
	"net/http"
	"ws-vpn-go/common"
)

type Tunnel interface {
	Run() error                                                 //Открываем тунель
	RegisterHandlers(mux *http.ServeMux) error                  //Вехаем хендлеры, необходимые для тунеля
	ReserveConnection(ip common.IpAddress) error                //Даем тунелю знать, что сервер выдал IP
	SetConnectionCloseHandler(handler func (common.IpAddress))  //Вызываем, чтобы отчистить выданный адрес, когда клиент отваливается
	WriteTo(target io.Writer) error                             //Читаем пакеты из тунеля и пишем в ...
	WriteToTunnel(target common.IpAddress, packet []byte) error //Пишем пакет в тунель
}