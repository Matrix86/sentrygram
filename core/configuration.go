package core

import (
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
	Logs        string   `json:"logs"`
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
