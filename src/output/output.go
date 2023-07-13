package output

import (
	"fmt"

	"github.com/lzfDream/ReadExcel/config"
	"github.com/lzfDream/ReadExcel/parse"
	"github.com/lzfDream/ReadExcel/types"
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
		fmt.Println(types.CustomTypeDetail)
		for fileName, defineFile := range types.CustomTypeDetail {
			err := OutputCSharpClassDefineFile(fileName, defineFile)
			if err != nil {
				return err
			}
		}
		err := OutputCSharp(cfg.OutputCSharpPath, defines, groupType)
		if err != nil {
			return err
		}
	}
	return nil
}
