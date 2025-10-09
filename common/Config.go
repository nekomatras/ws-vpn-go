package common

import (
	"encoding/json"
	"os"
	"fmt"
)

type Config struct {
	Mode             string `json:"mode"`               //Common
	RemoteAddress    string `json:"remote_address"`     //Client
	ListenAddress    string `json:"listen_address"`     //Server
	RegisterPath     string `json:"register_path"`      //Common
	TunnelType       string `json:"tunnel_type"`        //Common
	TunnelPath       string `json:"tunnel_path"`        //Common
	InterfaceName    string `json:"interface_name"`     //Common
	Key              string `json:"key"`                //Common
	MTU                uint `json:"mtu"`                //Server
	Network          string `json:"network"`            //Server
	DefaultPagePath  string `json:"default_page_path"`  //Server
	StaticFolderPath string `json:"static_folder_path"` //Server
}

var DefaultConfig = Config{
	Mode:             "client",
	RemoteAddress:    "127.0.0.1:443",
	ListenAddress:    "0.0.0.0:443",
	RegisterPath:     "/register",
	TunnelType:       "WS",
	TunnelPath:       "/ws",
	InterfaceName:    "tunWS",
	Key:              "SKEBOB",
	MTU:              1450,
	Network:          "10.0.0.0/24",
	DefaultPagePath:  "./index.html",
	StaticFolderPath: "./static/",
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("cannot open config file: %v", err)
	}
	defer file.Close()

	cfg := DefaultConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("cannot decode config: %v", err)
	}

	return &cfg, nil
}