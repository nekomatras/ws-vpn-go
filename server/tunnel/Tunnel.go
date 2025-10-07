package tunnel

import (
	"io"
	"net/http"
	"ws-vpn-go/common"
)

type Tunnel interface {
	Listen() error                                              //Открываем тунель
	RegisterHandlers(mux *http.ServeMux) error                  //Вехаем хендлеры, необходимые для тунеля
	WriteTo(target io.Writer) error                             //Читаем пакеты из тунеля и пишем в ...
	WriteToTunnel(target common.IpAddress, packet []byte) error //Пишем пакет в тунель
}