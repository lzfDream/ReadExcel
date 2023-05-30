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

func ParseCellValue(fieldDefine ExcelDefineField, cell string) (string, error) {
	if fieldDefine.Type == "bool" {
		b, err := strconv.ParseBool(cell)
		if err != nil {
			return "", err
		}
		if b {
			return `"` + fieldDefine.Name + `":true`, nil
		} else {
			return `"` + fieldDefine.Name + `":false`, nil
		}
	} else if fieldDefine.Type == "int" {
		num, err := strconv.Atoi(cell)
		if err != nil {
			return "", err
		}
		return `"` + fieldDefine.Name + `":` + strconv.Itoa(num), nil
	} else if fieldDefine.Type == "double" {
		num, err := strconv.ParseFloat(cell, 64)
		if err != nil {
			return "", err
		}
		return `"` + fieldDefine.Name + `":` + strconv.FormatFloat(num, 'f', -1, 64), nil
	} else if fieldDefine.Type == "string" {
		return `"` + fieldDefine.Name + `":"` + cell + `"`, nil
	}
	return "", fmt.Errorf("invalid field type %s", fieldDefine.Type)
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
	e.FileName = fileName
	e.SheetName = sheetName

	if len(rows) < 2 {
		return fmt.Errorf("%s, rows < 2", e.Desc())
	}

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

	if e.SheetFileType == SheetFileType_Horizontal {
		if len(rows) < 5 {
			return fmt.Errorf("%s, rows < 5", e.Desc())
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
	} else if e.SheetFileType == SheetFileType_Vertical {
		for _, row := range rows[3:] {
			if len(row) < 5 {
				return fmt.Errorf("%s, col size < 5", e.Desc())
			}
			field := ExcelDefineField{}
			field.Name = row[1]
			field.Type = row[2]
			if len(row) >= 3 && row[3] != "" {
				field.Group = row[3]
			} else {
				field.Group = "c,s"
			}
			e.Fields = append(e.Fields, field)
		}
	}

	return nil
}
