package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

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

			rowData := []string{}
			if define.SheetFileType == parse.SheetFileType_Horizontal {
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
						value, err := parse.ParseCellValue(fieldDefine, cell)
						if err != nil {
							fmt.Println(err)
							return
						}
						colData = append(colData, value)
					}
					rowData = append(rowData, "{"+strings.Join(colData, ",")+"}")
				}
			} else if define.SheetFileType == parse.SheetFileType_Vertical {
				colData := []string{}
				for index, row := range rows[3:] {
					fieldDefine := define.Fields[index]
					value, err := parse.ParseCellValue(fieldDefine, row[4])
					if err != nil {
						fmt.Println(err)
						return
					}
					colData = append(colData, value)
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
		elapsed := time.Since(startTime)
		fmt.Printf("读取%s完成, 耗时%s\n\n", fileName, elapsed)
	}
}
