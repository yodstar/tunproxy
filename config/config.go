package config

import (
	"encoding/json"
	"io/ioutil"
)

// Config
type Config struct {
	Listen  string
	Server  string
	Forward map[string]string
	Logger  struct {
		Outfile string
		Filter  string
		Level   string
	}
}

var CONF = &Config{}

// LoadFile
func LoadFile(path string) (err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, CONF)
}
