package parse

import (
	"fmt"
	"strconv"
)

const (
	SheetFileType_Vertical = iota
	SheetFileType_Horizontal
)

func getRowIndexByValue(row []string, value string) int {
	for index, v := range row {
		if v == value {
			return index
		}
	}
	return -1
}

func getRowValueByIndex(row []string, index int) string {
	if len(row) < index {
		return ""
	}
	return row[index]
}

type ExcelDefineField struct {
	Name  string
	Type  string
	Group string
}

type ExcelDefine struct {
	FileName      string
	SheetName     string
	OutFileName   string
	SheetFileType int
	Fields        []ExcelDefineField
}

func (e ExcelDefine) Desc() string {
	return e.FileName + ":" + e.SheetName
}

func (e *ExcelDefine) Parse(fileName, sheetName string, rows [][]string) error {
	if len(rows) < 5 {
		return fmt.Errorf("rows < 5")
	}
	e.FileName = fileName
	e.SheetName = sheetName

	index := getRowIndexByValue(rows[0], "OutFileName")
	if index == -1 {
		e.OutFileName = e.SheetName
	} else {
		e.OutFileName = getRowValueByIndex(rows[1], index)
	}

	index = getRowIndexByValue(rows[0], "SheetFileType")
	if index == -1 {
		e.SheetFileType = SheetFileType_Horizontal
	} else {
		num, err := strconv.Atoi(getRowValueByIndex(rows[1], index))
		if err != nil {
			return fmt.Errorf("%s, SheetFileType not int, err %w", e.Desc(), err)
		}
		e.SheetFileType = num
	}

	if len(rows[2]) < 1 {
		return fmt.Errorf("%s, rows line 3 is empty", e.Desc())
	}
	if len(rows[3]) < 1 {
		return fmt.Errorf("%s, rows line 4 is empty", e.Desc())
	}
	if len(rows[4]) < 1 {
		return fmt.Errorf("%s, rows line 5 is empty", e.Desc())
	}

	for i := 1; i < len(rows[2]); i++ {
		field := ExcelDefineField{}
		field.Name = rows[2][i]
		if i >= len(rows[3]) {
			return fmt.Errorf("%s, rows line 3 size < line 2 size", e.Desc())
		}
		field.Type = rows[3][i]
		if i < len(rows[4]) && rows[4][i] != "" {
			field.Group = rows[4][i]
		} else {
			field.Group = "c,s"
		}

		e.Fields = append(e.Fields, field)
	}
	return nil
}
