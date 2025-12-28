package common

import (
	"os"
	"fmt"
	"encoding/json"

	"github.com/creasty/defaults"
	"github.com/caarlos0/env/v9"
)

type Config struct {
	Mode             string `json:"mode"               env:"MODE"               default:"client"`                //Common
	Logger           string `json:"logger"             env:"LOGGER"             default:"INFO"`                  //Common
	RegisterPath     string `json:"register_path"      env:"REGISTER_PATH"      default:"/register"`             //Common
	TunnelType       string `json:"tunnel_type"        env:"TUNNEL_TYPE"        default:"WS"`                    //Common
	InterfaceName    string `json:"interface_name"     env:"INTERFACE_NAME"     default:"tunWS"`                 //Common
	Key              string `json:"key"                env:"KEY"                default:"SKEBOB"`                //Common
	RemoteAddress    string `json:"remote_address"     env:"REMOTE_ADDRESS"     default:"https://127.0.0.1"`     //Client
	ListenAddress    string `json:"listen_address"     env:"LISTEN_ADDRESS"     default:"0.0.0.0"`           //Server
	Secure           bool   `json:"secure"             env:"SECURE"             default:"true"`                  //Server
	Chain            string `json:"chain"              env:"CHAIN"              default:"./chain.pem"`           //Server
	PrivateKey       string `json:"private_key"        env:"PRIVATE_KEY"        default:"./key.pem"`             //Server
	TunnelPath       string `json:"tunnel_path"        env:"TUNNEL_PATH"        default:"/ws"`                   //Server
	MTU                uint `json:"mtu"                env:"MTU"                default:"1432"`                  //Server
	Network          string `json:"network"            env:"NETWORK"            default:"10.0.0.0/24"`           //Server
	DefaultPagePath  string `json:"default_page_path"  env:"DEFAULT_PAGE_PATH"  default:"./index.html"`          //Server
	StaticFolderPath string `json:"static_folder_path" env:"STATIC_FOLDER_PATH" default:"./static/"`             //Server
}

func LoadConfigFromFile(path string) (*Config, error) {
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

func LoadConfigFromEnvieronment() (*Config, error) {
	cfg := &Config{}

	if err := defaults.Set(cfg); err != nil {
        return nil, fmt.Errorf("cannot apply defaults: %v", err)
    }

	if err := env.Parse(cfg); err != nil {
        return nil, fmt.Errorf("cannot read env variables: %v", err)
    }

	return cfg, nil
}