package output

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/lzfDream/ReadExcel/parse"
	"github.com/sirupsen/logrus"
)

func FixCSharpType(strType string) string {
	switch strType {
	case "string", "bool", "int", "double":
		return strType
	case "string?", "bool?", "int?", "double?":
		return strings.TrimSuffix(strType, "?")
	}
	return ""
}

const TableText = `// 由github.com/lzfDream/ReadExcel生成, 请勿修改
using System;
using System.Collections.Generic;
using System.IO;
using System.Text.Json;

namespace Table
{
    public class Table{{.ClassName}} : ITableModule
    {
        public class Item
        {
            {{- range $_, $field := .Fields}}
            public {{$field.Type}} {{$field.Name}} { get; set; }
            {{- end}}
        }

        public {{range $_, $key := .Keys}}Dictionary<{{$key.Type}}, {{end}}Item{{range $_, $_ := .Keys}}>{{end}} AllItem;

        public void Load(in string path)
        {
            string json = File.ReadAllText(path + "/{{.FileName}}.json");
            AllItem = JsonSerializer.Deserialize<{{range $_, $key := .Keys}}Dictionary<{{$key.Type}}, {{end}}Item{{range $_, $_ := .Keys}}>{{end}}>(json);
        }

        public Item Get({{range $index, $key := .Keys}}{{if eq $index 1}}, {{end}}{{$key.Type}} {{$key.Name}}{{end}})
        {
            var dict0 = AllItem;
            {{- range $index, $key := .Keys}}
            if (!dict{{$index}}.TryGetValue({{$key.Name}}, out var dict{{add $index 1}}))
            {
                Debug.TableErrorLog(string.Format("get Table{{$.ClassName}} data fail, key {{$key.Name}}: {0}", {{$key.Name}}));
                return default;
            }
            {{- end}}
            return dict{{len .Keys}};
        }

        public {{range $_, $key := .Keys}}Dictionary<{{$key.Type}}, {{end}}Item{{range $_, $_ := .Keys}}>{{end}} GatAllItem()
        {
            return AllItem;
        }
    }
}
`

const TableText2 = `// 由github.com/lzfDream/ReadExcel生成, 请勿修改
using System;
using System.IO;
using System.Text.Json;

namespace Table
{
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
}
`

type TableKey struct {
	Name string
	Type string
}

type TableTemplate struct {
	FileName  string
	ClassName string
	Fields    []parse.ExcelDefineField
	Keys      []TableKey
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
	text, err := template.New(className).Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}).Parse(tempText)
	if err != nil {
		return err
	}
	data := TableTemplate{
		FileName:  define.OutFileName,
		ClassName: className,
	}
	for index, field := range define.Fields {
		if field.Group.HasGroup(groupType) {
			data.Fields = append(data.Fields, field)
		}
		if index < define.KeyNum {
			data.Keys = append(data.Keys, TableKey{field.Name, field.Type})
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

const TableMgrText = `// 由github.com/lzfDream/ReadExcel生成, 请勿修改
using System;
using System.Collections.Generic;

namespace Table
{
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
            tables = new List<ITableModule>
            {
                {{- range $_, $name := .}}
                new Table{{$name}}(),
                {{- end}}
            }
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
            return default;
        }
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
