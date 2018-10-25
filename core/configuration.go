package core

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/evilsocket/islazy/log"
)

type Configuration struct {
	Logs        string   `json:"logs"`
	LogsLevel   string   `json:"log_level"`
	TgmAPI      string   `json:"tgm_api"`
	PluginsPath string   `json:"plugins_path"`
	Users       []string `json:"users"`
	RpcEnabled  bool     `json:"rpc_enabled"`
	RpcPort     int      `json:"rpc_port"`
}

// Read a JSON conf file and return the Configuration object
func LoadConfiguration(path string) (Configuration, error) {
	configuration := Configuration{}

	file, err := os.Open(path)
	if err != nil {
		return configuration, fmt.Errorf("LoadConfiguration: file opening: %s", err)
	}

	err = json.NewDecoder(file).Decode(&configuration)
	if err != nil {
		return configuration, fmt.Errorf("LoadConfiguration: file decoding: %s", err)
	}

	return configuration, nil
}

func (c *Configuration) GetLogLevel() log.Verbosity {
	switch c.LogsLevel {
	case "debug":
		return log.DEBUG
	case "info":
		return log.INFO
	case "error":
		return log.ERROR

	default:
		return log.INFO
	}
}
