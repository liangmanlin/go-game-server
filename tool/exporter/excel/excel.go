package excel

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"path/filepath"
	"regexp"
	"strings"
)

var reg = regexp.MustCompile(``)

func LoadFile(fileName string) *Excel {
	file, err := excelize.OpenFile(fileName)
	if err != nil {
		panic(fmt.Errorf("读取文件：%s 失败：%s", fileName, err.Error()))
	}
	sheetName := file.GetSheetName(1)
	rows, err := file.GetRows(sheetName)
	if err != nil {
		panic(fmt.Errorf("读取文件：%s 失败：%s", fileName, err.Error()))
	}
	if len(rows) < 4 {
		return nil
	}
	heads := readHead(file, sheetName, rows)
	var all [][]string
	// 读取数据
	for i := 4; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 || row[0] == "" {
			continue
		}
		var dataRow []string
		for _, head := range heads {
			if len(row) <= head.Index {
				dataRow = append(dataRow, "")
				continue
			}
			val := row[head.Index]
			switch head.Type {
			case FTypeInt:
				if val == "" {val = "0"}
			case FTypeString:
				val = CheckValString(val)
			}
			dataRow = append(dataRow, val)
		}
		all = append(all, dataRow)
	}
	fileName = filepath.Base(fileName)
	ext := filepath.Ext(fileName)
	fileName = strings.TrimSuffix(fileName, ext)
	excel := &Excel{Name: fileName, Header: heads, AllData: all}
	return excel
}

func readHead(file *excelize.File, sheetName string, rows [][]string) []*Head {
	var h int = 'A'
	// 第二行是head定义
	row := rows[1]
	var heads []*Head
	for i, v := range row {
		rule := getRule(h, 2, file, sheetName)
		h++
		if isColSkip(rule, v) {
			continue
		}
		TypeName := strings.Trim(rows[2][i], " \t\n") // 类型定义
		var commit string
		if len(rows[3]) > i {
			commit = rows[3][i] // 注释
		}
		head := &Head{Rule: rule, Name: FirstUP(v), Type: getHType(TypeName), TypeName: TypeName, Commit: commit, Index: i}
		heads = append(heads, head)

	}
	return heads
}

func isColSkip(rule Rule, val string) bool {
	return rule == RuleNone || val == ""
}

func getRule(h int, index int, file *excelize.File, sheetName string) Rule {
	var axis string
	if h <= 'Z' {
		axis = fmt.Sprintf("%s%d", string(h), index)
	} else {
		h -= 'A'
		i := h%26 + 'A' // 这里假设不可能超过26*26个字段
		h = 'A' + h/26 - 1
		axis = fmt.Sprintf("%s%s%d", string(h), string(i), index)
	}
	styleID, _ := file.GetCellStyle(sheetName, axis)
	fillID := file.Styles.CellXfs.Xf[styleID].FillID
	fgColor := file.Styles.Fills.Fill[fillID].PatternFill.FgColor
	return ruleMap[fgColor.Indexed]
}

func getHType(TypeName string) FType {
	switch TypeName {
	case "int":
		return FTypeInt
	case "string":
		return FTypeString
	case "key":
		return FTypeKey
	default:
		return FTypeTerm
	}
}

func CheckValString(val string) string {
	if len(val) == 0 {
		return "\"\""
	}
	if val[0] == '"' {
		return val
	}
	return "\"" + val + "\""
}

func FirstUP(val string) string {
	if val[0] >= 'A' && val[0] <= 'Z' {
		return val
	}
	return string(val[0]-('a'-'A')) + val[1:]
}