package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/lzfDream/ReadExcel/config"
	"github.com/lzfDream/ReadExcel/parse"

	"github.com/xuri/excelize/v2"
)

func main() {
	begin := time.Now()
	cfg := config.GetConfig()

	entries, err := os.ReadDir(cfg.InputPath)
	if err != nil {
		fmt.Println("读取目录失败：", err)
		return
	}
	defines := make([]parse.ExcelDefine, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		fmt.Println("开始读取文件: ", fileName)
		startTime := time.Now()

		f, err := excelize.OpenFile(cfg.InputPath + "/" + fileName)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer func() {
			if err := f.Close(); err != nil {
				fmt.Println(err)
				return
			}
		}()

		sheets := f.GetSheetList()
		for _, sheetName := range sheets {
			rows, err := f.GetRows(sheetName)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("开始解析 %s:%s\n", fileName, sheetName)
			define := parse.ExcelDefine{}
			err = define.Parse(fileName, sheetName, rows)
			if err != nil {
				fmt.Println(err)
				return
			}
			defines = append(defines, define)

			clientData := map[string]interface{}{}
			if define.SheetFileType == parse.SheetFileType_Horizontal {
				for _, row := range rows[5:] {
					data := clientData
					// 空行停止
					if len(row) == 0 {
						break
					}
					// 注释行
					if strings.HasPrefix(row[0], "##") {
						continue
					}
					colData := map[string]interface{}{}
					for index, cell := range row[1:] {
						fieldDefine := define.Fields[index]
						value, err := parse.ParseCellValue(fieldDefine, cell)
						if err != nil {
							fmt.Println(err)
							return
						}
						if index < define.KeyNum {
							key := cell
							data[key] = map[string]interface{}{}
							data = data[key].(map[string]interface{})
						}
						if fieldDefine.Group.HasGroup(parse.GroupType_Client) {
							colData[fieldDefine.Name] = value
						}
					}
					for k, v := range colData {
						data[k] = v
					}
				}
			} else if define.SheetFileType == parse.SheetFileType_Vertical {
				for index, row := range rows[3:] {
					fieldDefine := define.Fields[index]
					value, err := parse.ParseCellValue(fieldDefine, row[4])
					if err != nil {
						fmt.Println(err)
						return
					}
					if fieldDefine.Group.HasGroup(parse.GroupType_Client) {
						clientData[fieldDefine.Name] = value
					}
				}
			}
			if len(clientData) == 0 {
				fmt.Printf("%s:%s client 数据为空, 被跳过\n", fileName, sheetName)
			} else {
				if cfg.Client.OutputFileType == config.OutputFileType_Json {
					err = OutputJson(cfg.Client.OutputPath+"/"+define.OutFileName, clientData)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
			}
		}
		elapsed := time.Since(startTime)
		fmt.Printf("读取%s完成, 耗时%s\n\n", fileName, elapsed)
	}
	OutputCSharp(defines)
	elapsed := time.Since(begin)
	fmt.Printf("总耗时%s\n", elapsed)
}

func OutputJson(fileName string, data map[string]interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	jsonData = append(jsonData, '\n')
	err = ioutil.WriteFile(fileName+".json", jsonData, 0644)
	if err != nil {
		return fmt.Errorf("写入文件%s失败: %w", fileName, err)
	}

	return nil
}

func FixCSharpType(strType string) string {
	switch strType {
	case "string", "bool", "int", "double":
		return strType
	case "string?", "bool?", "int?", "double?":
		return strings.TrimSuffix(strType, "?")
	}
	return ""
}

const TableText = `// 生成
using System;
using System.Collections;
using System.Collections.Generic;

struct {{.ClassName}}
{
    {{- range _, $field := .}}
    public {{$field.Type}} {{$field.Name}};
    {{- end}}
}
`

func OutputCSharpCode(path string, define parse.ExcelDefine) error {
	lines := make([]string, 0, 10)
	if define.SheetFileType == parse.SheetFileType_Vertical {
		text, err := template.New("test").Parse(TableText)
		if err != nil {
			return nil
		}

		var buf strings.Builder
		err = text.Execute(&buf, define.Fields)
		if err != nil {
			return nil
		}
		lines = append(lines, buf.String())
	}
	str := strings.Join(lines, "\n")
	filePath := path + define.OutFileName + ".cs"
	err := ioutil.WriteFile(filePath, []byte(str), 0644)
	if err != nil {
		return fmt.Errorf("写入文件%s失败: %w", filePath, err)
	}

	return nil
}

const TableMgrText = `// 生成
using Table;
using System.Collections.Generic;

interface ITableModule {
    void load();
}

public class TableMgr {
    private List<ITableModule> tables;
    private static TableMgr instance;

    private TableMgr() {
        {{- range $index, $ClassName := .}}
        tables.Add(new {{$ClassName}}())
        {{- end}}
    }

    public static TableMgr Instance() {
        if (instance == null) {
            instance = new TableMgr();
        }
        return instance;
    }

    public void load() {
        foreach(ITableModule table in tables) {
            table.load();
        }
    }
}
`

func OutputCSharp(defines []parse.ExcelDefine) error {
	// 生成每一个sheet的类
	for _, define := range defines {
		err := OutputCSharpCode("code/", define)
		if err != nil {
			return err
		}
	}

	// 生成全局管理类
	text, err := template.New("test").Parse(TableMgrText)
	if err != nil {
		return nil
	}

	classNames := make([]string, 0)
	for _, define := range defines {
		classNames = append(classNames, define.OutFileName)
	}
	var buf strings.Builder
	err = text.Execute(&buf, classNames)
	if err != nil {
		return nil
	}

	lines := make([]string, 0)
	lines = append(lines, buf.String())
	str := strings.Join(lines, "\n")
	filePath := "code/TableMgr.cs"
	err = ioutil.WriteFile(filePath, []byte(str), 0644)
	if err != nil {
		return fmt.Errorf("写入文件%s失败: %w", filePath, err)
	}
	return nil
}
