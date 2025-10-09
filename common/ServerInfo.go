package common

import (
	"encoding/json"
	"net/http"
)

type ServerInfo struct {
	MTU                   uint
	InternalServerAddress string

	ClientIp              string
}

func (info ServerInfo) WriteToResponse(w http.ResponseWriter) {
    json.NewEncoder(w).Encode(info)
}