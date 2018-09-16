package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	ServerPort       string `json:"server_port"`
	ConnStr          string `json:"connstr"`
	MaxDBConnections int    `json:"max_db_connections"`
	TemplateDir      string `json:"template_dir"`
}

func NewConfig(configFileName string) (*Config, error) {
	file, err := os.Open(configFileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	c := &Config{}
	dec := json.NewDecoder(file)
	err = dec.Decode(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
