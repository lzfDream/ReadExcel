package main

import (
	"fmt"
	"io/ioutil"
	"table/config"

	"github.com/xuri/excelize/v2"
)

func main() {
	cfg := config.GetConfig()

	f, _ := excelize.OpenFile(cfg.InputPath + "/item.xlsx")
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheets := f.GetSheetList()
	str := ""
	for _, sheetName := range sheets {
		rows, err := f.GetRows(sheetName)
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, row := range rows {
			for _, colCell := range row {
				str += colCell + "\t"
			}
			str += "\n"
		}
	}

	err := ioutil.WriteFile(cfg.OutputPath+"/test.txt", []byte(str), 0644)
	if err != nil {
		fmt.Println("写入文件失败：", err)
		return
	}
}
