package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/lzfDream/ReadExcel/config"
	"github.com/lzfDream/ReadExcel/parse"

	"github.com/xuri/excelize/v2"
)

func main() {
	cfg := config.GetConfig()

	entries, err := os.ReadDir(cfg.InputPath)
	if err != nil {
		fmt.Println("读取目录失败：", err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		f, err := excelize.OpenFile(cfg.InputPath + "/" + entry.Name())
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
			define := parse.ExcelDefine{}
			err = define.Parse(entry.Name(), sheetName, rows)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("%+v\n", define)

			rowData := []string{}
			for _, row := range rows[5:] {
				// 空行停止
				if len(row) == 0 {
					break
				}
				// 注释行
				if strings.HasPrefix(row[0], "##") {
					continue
				}
				colData := []string{}
				for index, cell := range row[1:] {
					fieldDefine := define.Fields[index]
					if fieldDefine.Type == "bool" {
						b, err := strconv.ParseBool(cell)
						if err != nil {
							fmt.Println(err)
							return
						}
						if b {
							colData = append(colData, `"`+fieldDefine.Name+`":true`)
						} else {
							colData = append(colData, `"`+fieldDefine.Name+`":false`)
						}
					} else if fieldDefine.Type == "int" {
						num, err := strconv.Atoi(cell)
						if err != nil {
							fmt.Println(err)
							return
						}
						colData = append(colData, `"`+fieldDefine.Name+`":`+strconv.Itoa(num))
					} else if fieldDefine.Type == "double" {
						num, err := strconv.ParseFloat(cell, 64)
						if err != nil {
							fmt.Println(err)
							return
						}
						colData = append(colData, `"`+fieldDefine.Name+`":`+strconv.FormatFloat(num, 'f', -1, 64))
					} else if fieldDefine.Type == "string" {
						colData = append(colData, `"`+fieldDefine.Name+`":"`+cell+`"`)
					}
				}
				rowData = append(rowData, "{"+strings.Join(colData, ",")+"}")
			}
			str := "[" + strings.Join(rowData, ",") + "]"
			var out bytes.Buffer
			json.Indent(&out, []byte(str), "", "    ")
			err = ioutil.WriteFile(cfg.Client.OutputPath+"/"+define.OutFileName+".json", out.Bytes(), 0644)
			if err != nil {
				fmt.Println("写入文件失败：", err)
				return
			}
		}
		break
	}
}
