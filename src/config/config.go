package config

import (
	"encoding/json"
	"io/ioutil"
)

const (
	OutputDataType_Json = "json"
)

const (
	OutputCodeType_CSharp = "c#"
)

type GroupConfig struct {
	OutputCSharpPath string `json:"output_csharp_path,omitempty"`
	OutputPath       string `json:"output_path,omitempty"`
	OutputDataType   string `json:"output_data_type,omitempty"`
	OutputCodeType   string `json:"output_code_type,omitempty"`
}

type Config struct {
	InputPath string      `json:"input_path,omitempty"`
	Client    GroupConfig `json:"client,omitempty"`
}

var cfg Config

func init() {
	bytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic("read config.json fail, err: " + err.Error())
	}

	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		panic("config.json Unmarshal fail, err: " + err.Error())
	}
	if cfg.Client.OutputDataType == "" {
		cfg.Client.OutputDataType = OutputDataType_Json
	}
	if cfg.Client.OutputCodeType == "" {
		cfg.Client.OutputCodeType = OutputCodeType_CSharp
	}
}

func GetConfig() Config {
	return cfg
}
