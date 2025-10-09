package common

import (
	"encoding/json"
	"os"
	"fmt"
)

type Config struct {
	Mode             string `json:"mode"`            //Common
	RemoteAddress    string `json:"remote_address"`  //Client
	ListenAddress    string `json:"listen_address"`  //Server
	RegisterPath     string `json:"register_path"`   //Common
	TunnelType       string `json:"tunnel_type"`     //Common???
	TunnelPath       string `json:"tunnel_path"`     //Common
	InterfaceName    string `json:"interface_name"`  //Common
	Key              string `json:"key"`             //Common
	MTU                uint `json:"mtu"`             //Server
	Network          string `json:"network"`         //Server
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("cannot open config file: %v", err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("cannot decode config: %v", err)
	}

	return &cfg, nil
}