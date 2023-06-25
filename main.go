package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/lzfDream/ReadExcel/config"
	"github.com/lzfDream/ReadExcel/parse"

	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	begin := time.Now()
	cfg := config.GetConfig()

	entries, err := os.ReadDir(cfg.InputPath)
	if err != nil {
		logrus.Errorln("读取目录失败：", err)
		return
	}
	defines := make([]parse.ExcelDefine, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		logrus.Infof("开始读取文件: %s", fileName)
		startTime := time.Now()

		f, err := excelize.OpenFile(cfg.InputPath + "/" + fileName)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer func() {
			if err := f.Close(); err != nil {
				logrus.Errorln(err)
				return
			}
		}()

		sheets := f.GetSheetList()
		for _, sheetName := range sheets {
			rows, err := f.GetRows(sheetName)
			if err != nil {
				logrus.Errorln(err)
				return
			}
			logrus.Infof("开始解析 %s:%s", fileName, sheetName)
			define := parse.ExcelDefine{}
			err = define.Parse(fileName, sheetName, rows)
			if err != nil {
				logrus.Errorln(err)
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
				logrus.Infof("%s:%s client 数据为空, 被跳过", fileName, sheetName)
			} else {
				if cfg.Client.OutputFileType == config.OutputFileType_Json {
					err = OutputJson(cfg.Client.OutputPath, define.OutFileName, clientData)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
			}
		}
		elapsed := time.Since(startTime)
		logrus.Infof("读取%s完成, 耗时%s\n\n", fileName, elapsed)
	}
	err = OutputCSharp(cfg.Client.OutputCSharpPath, defines, parse.GroupType_Client)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	logrus.Infof("导出成功, 总耗时%s", time.Since(begin))
}

func OutputJson(path, fileName string, data map[string]interface{}) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	jsonData = append(jsonData, '\n')
	err = os.WriteFile(path+"/"+fileName+".json", jsonData, 0644)
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
using System.Collections.Generic;
using System.IO;
using System.Text.Json;

public class Table{{.ClassName}} : ITableModule
{
    public class Item
    {
        {{- range $_, $field := .Fields}}
        public {{$field.Type}} {{$field.Name}} { get; set; }
        {{- end}}
    }

    public Dictionary<string, Item> AllItem;

    public void Load(in string path)
    {
        string json = File.ReadAllText(path + "/{{.FileName}}.json");
        AllItem = JsonSerializer.Deserialize<Dictionary<string, Item>>(json);
    }
}
`

const TableText2 = `// 生成
using System;
using System.IO;
using System.Text.Json;

public class Table{{.ClassName}} : ITableModule
{
    public class Item
    {
        {{- range $_, $field := .Fields}}
        public {{$field.Type}} {{$field.Name}} { get; set; }
        {{- end}}
    }

    public Item KeyItem;

    public void Load(in string path)
    {
        string json = File.ReadAllText(path + "/{{.FileName}}.json");
        KeyItem = JsonSerializer.Deserialize<Item>(json);
    }
}
`

type TableTemplate struct {
	FileName  string
	ClassName string
	Fields    []parse.ExcelDefineField
}

func OutputCSharpCode(path string, define parse.ExcelDefine, groupType parse.GroupType) error {
	className := define.OutFileName
	className = parse.CaseToCamel(className)
	begin := time.Now()
	defer func() {
		logrus.Infof("输出c#类%s耗时%s", className, time.Since(begin))
	}()

	tempText := TableText
	if define.SheetFileType == parse.SheetFileType_Vertical {
		tempText = TableText2
	}
	str := ""
	text, err := template.New(className).Parse(tempText)
	if err != nil {
		return err
	}
	data := TableTemplate{
		FileName:  define.OutFileName,
		ClassName: className,
	}
	for _, field := range define.Fields {
		if field.Group.HasGroup(groupType) {
			data.Fields = append(data.Fields, field)
		}
	}

	var buf strings.Builder
	err = text.Execute(&buf, data)
	if err != nil {
		return err
	}
	str = buf.String()
	filePath := path + "/" + define.OutFileName + ".cs"
	err = os.WriteFile(filePath, []byte(str), 0644)
	if err != nil {
		return fmt.Errorf("写入文件%s失败: %w", filePath, err)
	}

	return nil
}

const TableMgrText = `// 生成
using Table;
using System.Collections.Generic;

interface ITableModule
{
    void Load(in string path);
}

public class TableMgr
{
    private List<ITableModule> tables;
    private static TableMgr instance;

    private TableMgr()
    {
        tables = new List<ITableModule>();
        {{- range $_, $name := .}}
        tables.Add(new Table{{$name}}());
        {{- end}}
    }

    public static TableMgr Instance()
    {
        if (instance == null)
        {
            instance = new TableMgr();
        }
        return instance;
    }

    public void Load(in string path)
    {
        foreach(ITableModule table in tables)
        {
            table.Load(path);
        }
    }

    public T GetTable<T>() where T : class
    {
        Type type = typeof(T);
        foreach(ITableModule table in tables)
        {
            Type type2 = table.GetType();
            if (type == type2)
            {
                return table as T;
            }
        }
        return default(T);
    }
}
`

func OutputCSharp(path string, defines []parse.ExcelDefine, groupType parse.GroupType) error {
	logrus.Infof("开始输出c#代码定义")
	begin := time.Now()
	defer func() {
		logrus.Infof("输出代码定义耗时%s\n\n", time.Since(begin))
	}()
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	// 生成每一个sheet的类
	classNames := []string{}
	for _, define := range defines {
		err := OutputCSharpCode(path, define, groupType)
		if err != nil {
			return err
		}
		classNames = append(classNames, parse.CaseToCamel(define.OutFileName))
	}

	mgrClassName := "TableMgr"
	begin2 := time.Now()
	defer func() {
		logrus.Infof("输出c#管理类%s耗时%s", mgrClassName, time.Since(begin2))
	}()
	// 生成全局管理类
	text, err := template.New(mgrClassName).Parse(TableMgrText)
	if err != nil {
		return nil
	}

	var buf strings.Builder
	err = text.Execute(&buf, classNames)
	if err != nil {
		return err
	}

	str := buf.String()
	filePath := path + "/" + mgrClassName + ".cs"
	err = os.WriteFile(filePath, []byte(str), 0644)
	if err != nil {
		return fmt.Errorf("写入文件%s失败: %w", filePath, err)
	}
	return nil
}
