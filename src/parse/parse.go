package parse

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lzfDream/ReadExcel/types"
	"github.com/xuri/excelize/v2"
)

const (
	SheetFileType_Vertical = iota
	SheetFileType_Horizontal
)

type GroupType int

const (
	GroupType_Client GroupType = 1 << iota
	GroupType_Server
	GroupType_All
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

func ParseCellValue(fieldDefine ExcelDefineField, cell string) (interface{}, error) {
	if !strings.HasSuffix(fieldDefine.Type, "?") {
		if cell == "" {
			return nil, fmt.Errorf("cell is empty")
		}
	}

	if cell == "" {
		var defVal interface{} = ""
		if fieldDefine.Type == "bool?" {
			defVal = false
		} else if fieldDefine.Type == "int?" {
			defVal = 0
		} else if fieldDefine.Type == "double?" {
			defVal = 0.0
		}
		return defVal, nil
	}

	if fieldDefine.Type == "bool" || fieldDefine.Type == "bool?" {
		b, err := strconv.ParseBool(cell)
		if err != nil {
			return nil, err
		}
		return b, nil
	} else if fieldDefine.Type == "int" || fieldDefine.Type == "int?" {
		num, err := strconv.Atoi(cell)
		if err != nil {
			return nil, err
		}
		return num, nil
	} else if fieldDefine.Type == "double" || fieldDefine.Type == "double?" {
		num, err := strconv.ParseFloat(cell, 64)
		if err != nil {
			return nil, err
		}
		return num, nil
	} else if fieldDefine.Type == "string" || fieldDefine.Type == "string?" {
		return cell, nil
	} else if classDefine, ok := types.CustomType[fieldDefine.Type]; ok {
		v := strings.Split(cell, classDefine.Separator)
		if len(v) != len(classDefine.Fields) {
			return nil, errors.New("field num !=")
		}
		data := make(map[string]interface{})
		for index, field := range classDefine.Fields {
			newFieldDefine := ExcelDefineField{
				field.Name,
				field.Type,
				fieldDefine.Group,
			}
			v, err := ParseCellValue(newFieldDefine, v[index])
			if err != nil {
				return nil, err
			}
			data[field.Name] = v
		}
		return data, nil
	}
	panic(fmt.Errorf("invalid field type %s", fieldDefine.Type))
}

func (g *GroupType) Parse(str string) {
	if len(str) == 0 {
		*g = GroupType_All
		return
	}
	groups := strings.Split(str, ",")
	for _, group := range groups {
		if group == "c" {
			*g |= GroupType_Client
		} else if group == "s" {
			*g |= GroupType_Server
		}
	}
}

func (g GroupType) HasGroup(group GroupType) bool {
	if g == GroupType_All {
		return true
	}
	return g&group == group
}

type ExcelDefineField struct {
	Name  string
	Type  string
	Group GroupType
}

type ExcelDefine struct {
	FileName      string
	SheetName     string
	OutFileName   string
	SheetFileType int
	Fields        []ExcelDefineField
	KeyNum        int
}

func (e ExcelDefine) Desc() string {
	return e.FileName + ":" + e.SheetName
}

func (e *ExcelDefine) Parse(fileName, sheetName string, rows [][]string) error {
	e.FileName = fileName
	e.SheetName = sheetName
	e.KeyNum = 1

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

	index = getRowIndexByValue(rows[0], "KeyNum")
	if index != -1 {
		num, err := strconv.Atoi(getRowValueByIndex(rows[1], index))
		if err != nil {
			return fmt.Errorf("%s, KeyNum not int, err %w", e.Desc(), err)
		}
		e.KeyNum = num
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
			field.Name = CaseToCamel(field.Name)
			if i >= len(rows[3]) {
				return fmt.Errorf("%s, rows line 3 size < line 2 size", e.Desc())
			}
			field.Type = rows[3][i]
			if i < len(rows[4]) {
				field.Group.Parse(rows[4][i])
			} else {
				field.Group = GroupType_All
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
			if len(row) >= 3 {
				field.Group.Parse(row[3])
			} else {
				field.Group = GroupType_All
			}
			e.Fields = append(e.Fields, field)
		}
	}

	for index, field := range e.Fields {
		if field.Name == "" {
			return fmt.Errorf("%s, field %d name is empty", e.Desc(), index)
		}
		if field.Type == "" {
			return fmt.Errorf("%s, field %d type is empty", e.Desc(), index)
		}
		if !types.IsValid(field.Type) {
			return fmt.Errorf("%s, field %d invalid type", e.Desc(), index)
		}
	}

	return nil
}

func CaseToCamel(name string) string {
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.Title(name)
	return strings.ReplaceAll(name, " ", "")
}

func ReadSheet(fileName, sheetName string, f *excelize.File) (ExcelDefine, map[string]interface{}, error) {
	define := ExcelDefine{}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return define, nil, err
	}

	err = define.Parse(fileName, sheetName, rows)
	if err != nil {
		return define, nil, err
	}

	clientData := map[string]interface{}{}
	if define.SheetFileType == SheetFileType_Horizontal {
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
				value, err := ParseCellValue(fieldDefine, cell)
				if err != nil {
					return define, nil, err
				}
				if index < define.KeyNum {
					key := cell
					data[key] = map[string]interface{}{}
					data = data[key].(map[string]interface{})
				}
				if fieldDefine.Group.HasGroup(GroupType_Client) {
					colData[fieldDefine.Name] = value
				}
			}
			for k, v := range colData {
				data[k] = v
			}
		}
	} else if define.SheetFileType == SheetFileType_Vertical {
		for index, row := range rows[3:] {
			fieldDefine := define.Fields[index]
			value, err := ParseCellValue(fieldDefine, row[4])
			if err != nil {
				return define, nil, err
			}
			if fieldDefine.Group.HasGroup(GroupType_Client) {
				clientData[fieldDefine.Name] = value
			}
		}
	}

	return define, clientData, nil
}
