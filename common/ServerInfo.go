package common

import (
	"encoding/json"
	"net/http"
)

type ServerInfo struct {
	MTU                   uint
	GatewayIp             string
	TunnelPath            string
	ClientIp              string
}

func (info ServerInfo) WriteToResponse(w http.ResponseWriter) {
    json.NewEncoder(w).Encode(info)
}