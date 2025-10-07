package common

import (
	"encoding/json"
	"os"
	"fmt"
)

type Config struct {
	Mode             string `json:"mode"`
	RemoteUrl        string `json:"remote_url"`
	InterfaceName    string `json:"interface_name"`
	//InterfaceAddress string `json:"interface_address"`
	Key              string `json:"key"`
	MTU                uint `json:"mtu"`
	Network          string `json:"network"`
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