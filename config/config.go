package config

import (
	"encoding/json"
	"io/ioutil"
)

type config struct {
	InputPath  string `json:"input_path"`
	OutputPath string `json:"output_path"`
}

var cfg config

func init() {
	bytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic("read config.json fail, err: " + err.Error())
	}

	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		panic("config.json Unmarshal fail, err: " + err.Error())
	}
}

func GetConfig() config {
	return cfg
}
