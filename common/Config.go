package common

import (
	"os"
	"fmt"
	"encoding/json"

	"github.com/creasty/defaults"
)

type Config struct {
	Mode             string `json:"mode"               default:"client"`        //Common
	RemoteAddress    string `json:"remote_address"     default:"127.0.0.1:443"` //Client
	ListenAddress    string `json:"listen_address"     default:"0.0.0.0:443"`   //Server
	RegisterPath     string `json:"register_path"      default:"/register"`     //Common
	TunnelType       string `json:"tunnel_type"        default:"WS"`            //Common
	TunnelPath       string `json:"tunnel_path"        default:"/ws"`           //Common
	InterfaceName    string `json:"interface_name"     default:"tunWS"`         //Common
	Key              string `json:"key"                default:"SKEBOB"`        //Common
	MTU                uint `json:"mtu"                default:"1450"`          //Server
	Network          string `json:"network"            default:"10.0.0.0/24"`   //Server
	DefaultPagePath  string `json:"default_page_path"  default:"./index.html"`  //Server
	StaticFolderPath string `json:"static_folder_path" default:"./static/"`     //Server
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)

	cfg := &Config{}

	if err := defaults.Set(cfg); err != nil {
        return nil, fmt.Errorf("cannot apply defaults: %v", err)
    }

	if err != nil {
		return nil, fmt.Errorf("cannot open config file: %v", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(cfg); err != nil {
		return nil, fmt.Errorf("cannot decode config: %v", err)
	}

	return cfg, nil
}