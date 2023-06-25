package config

import (
	"encoding/json"
	"io/ioutil"
)

const (
	OutputFileType_Json = "json"
)

type groupConfig struct {
	OutputCSharpPath string `json:"output_csharp_path,omitempty"`
	OutputPath       string `json:"output_path,omitempty"`
	OutputFileType   string `json:"output_file_type,omitempty"`
}

type config struct {
	InputPath string      `json:"input_path,omitempty"`
	Client    groupConfig `json:"client,omitempty"`
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
	if cfg.Client.OutputFileType == "" {
		cfg.Client.OutputFileType = OutputFileType_Json
	}
}

func GetConfig() config {
	return cfg
}
