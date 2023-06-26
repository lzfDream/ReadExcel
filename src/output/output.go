package output

import (
	"github.com/lzfDream/ReadExcel/config"
	"github.com/lzfDream/ReadExcel/parse"
)

func OutputData(cfg config.GroupConfig, define parse.ExcelDefine, data map[string]interface{}) error {
	if cfg.OutputDataType == config.OutputDataType_Json {
		err := OutputJson(cfg.OutputPath, define.OutFileName, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func OutputCode(cfg config.GroupConfig, defines []parse.ExcelDefine, groupType parse.GroupType) error {
	if cfg.OutputCodeType == config.OutputCodeType_CSharp {
		err := OutputCSharp(cfg.OutputCSharpPath, defines, groupType)
		if err != nil {
			return err
		}
	}
	return nil
}
