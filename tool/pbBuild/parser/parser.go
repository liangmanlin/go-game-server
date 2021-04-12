package parser

import (
	"fmt"
	"regexp"
	"strings"
	"unsafe"
)

var exp = regexp.MustCompile(`[\t ]*map[\t ]*\[[\t ]*\w+[\t ]*\][\t ]*\**[\t ]*\w+|[\t ]*\[[\t ]*\][\t ]*\*?[\t ]*\w+|[\t ]*\*?[\t ]*\w+`)
var fieldExp = regexp.MustCompile(`(map)*\[(\w+)\]\**(\w+)|\[\]\*?(\w+)|\*?(\w+)`)
var typeExp = regexp.MustCompile(`(\w+)(Tos|Toc)(\w+)`)

var tMap = map[string]PType{
	"bool":    PTBool,
	"int8":    PTInt8,
	"int16":   PTInt16,
	"int32":   PTInt32,
	"int64":   PTInt64,
	"uint16":  PTUInt16,
	"float32": PTFloat32,
	"float64": PTFloat64,
	"string":  PTString,
}

var TypeNameMap = map[PType]string{
	PTBool:    "bool",
	PTInt8:    "int8",
	PTInt16:   "int16",
	PTInt32:   "int32",
	PTInt64:   "int64",
	PTUInt16:  "uint16",
	PTFloat32: "float32",
	PTFloat64: "float64",
	PTString:  "string",
}

/*
根据语法规则，逐字分析
*/

func Scan(buf []byte) []*Proto {
	lens := len(buf)
	var all []*Proto
	modIndex := 0
	modMap := make(map[string]int)
	subModMap := make(map[string]int)
	for i := 0; i < lens; {
		i, modIndex, all = readStruct(buf, i, lens, all, modIndex, modMap, subModMap)
	}
	return all
}

func readStruct(buf []byte, index int, lens int, all []*Proto,
	modIndex int, modMap map[string]int, subModMap map[string]int) (int, int, []*Proto) {
	var ok, lineEnd bool
	var name, tmp, headC, lineC string
	ti := index
	index, headC = readHeadC(buf, lens, index)
	if index, ok = startWith("type", buf, lens, index); ok {
		index, name, _ = readWord(buf, lens, index)
		index, tmp, _ = readWord(buf, lens, index)
		if tmp != "struct" {
			return index, modIndex, all
		}
		index, tmp, lineEnd = readWord(buf, lens, index)
		if tmp == "" {
			lineC = ""
			index, tmp, lineEnd = readWord(buf, lens, index)
			if tmp != "{" {
				return index, modIndex, all
			}
			index = readLine(buf, lens, index)
			goto step1 //这里需要跳过读取行注释
		} else if tmp != "{" {
			return index, modIndex, all
		}
		if lineEnd {
			lineC = ""
		} else {
			index, lineC = readLineC(buf, lens, index) //行注释,会把换行符读取掉
		}
	step1:
		var fields []*Field
		for index < lens {
			index, fields, ok = readField(buf, lens, index, fields)
			if !ok {
				break
			}
		}
		pbType, mod, subMod := getPBType(name)
		var protoID int
		if pbType == PBTos || pbType == PBToc {
			protoID, modIndex = makeProtoID(modIndex, mod, subMod, modMap, subModMap)
		}
		p := &Proto{Name: name, Type: pbType, Mod: mod, SubMod: subMod, ProtoID: protoID, HeadC: headC, LineC: lineC, Fields: fields}
		return index, modIndex, append(all, p)
	}
	index = readLine(buf, lens, ti)
	return index, modIndex, all
}

func startWith(str string, buf []byte, lens int, index int) (int, bool) {
	i := 0
	b := *(*[]byte)(unsafe.Pointer(&str))
	strLen := len(b)
	for index < lens {
		v := buf[index]
		index++
		if v == 32 || v == 9 {
			continue
		}
		if v == b[i] {
			i++
			if i >= strLen {
				return index, true
			}
		} else {
			return index, false
		}
	}
	return index, false
}

func readWord(buf []byte, lens int, index int) (int, string, bool) {
	start := false
	startIndex := 0
	for index < lens {
		v := buf[index]
		index++
		if v == 32 || v == 9 {
			if !start {
				continue
			}
			return index, string(buf[startIndex : index-1]), false
		} else if v == 10 || v == 13 {
			if start {
				return index, string(buf[startIndex : index-1]), true
			}
			return index, "", true
		}
		if !start {
			start = true
			startIndex = index - 1
		}
	}
	return index, "", true
}

func readHeadC(buf []byte, lens, index int) (int, string) {
	var s []string
	var ok bool
step:
	tmp := index
	if index, ok = startWith("//", buf, lens, index); ok {
		start := index
		index = readLine(buf, lens, index)
		s = append(s, string(buf[start:index-1]))
		tmp = index
		goto step
	}
	return tmp, strings.Join(s, "\n")
}

func readLineC(buf []byte, lens, index int) (int, string) {
	for index < lens {
		v := buf[index]
		index++
		if v == 47 && buf[index] == 47 {
			index++
			start := index
			index = readLine(buf, lens, index)
			return index, string(buf[start : index-1])
		} else if v == 10 || v == 13 {
			return index, ""
		}
	}
	return index, ""
}

func readLine(buf []byte, lens int, index int) int {
	for index < lens {
		v := buf[index]
		index++
		if v == 10 || v == 13 {
			break
		}
	}
	return index
}

func readField(buf []byte, lens, index int, fields []*Field) (int, []*Field, bool) {
	var name, lineC string
	var fType *FieldType
	var lineEnd bool
	for index < lens {
		index, name, lineEnd = readWord(buf, lens, index)
		if name == "}" {
			return index, fields, false
		} else if name == "" {
			continue
		} else if _, ok := startWith("//", *(*[]byte)(unsafe.Pointer(&name)), len(name), 0); ok {
			index = readLine(buf, lens, index)
			continue
		}
		index, fType, lineEnd = readType(buf, lens, index)
		lineC = ""
		if !lineEnd {
			index, lineC = readLineC(buf, lens, index)
		}
		f := &Field{Name: name, FType: fType, LineC: lineC}
		return index, append(fields, f), true
	}
	return index, fields, false
}

func readType(buf []byte, lens, index int) (int, *FieldType, bool) {
	var name string
	b := exp.Find(buf[index:])
	index += len(b)
	i := 0
	// 去除空白
	for _, v := range b {
		if v == 32 || v == 9 {
			continue
		}
		b[i] = v
		i++
	}
	b = b[0:i]
	name = *(*string)(unsafe.Pointer(&b))
	s := fieldExp.FindStringSubmatch(name)
	var ft *FieldType
	if t := s[5]; t != "" {
		ft = makeFieldType(name, t)
	} else if s[1] != "" {
		ft = makeFieldTypeMap(name, s[2], s[3])
	} else {
		ft = makeFieldTypeSlice(name, s[4])
	}
	// 分析字段类型
	return index, ft, false
}

func makeFieldType(name, TName string) *FieldType {
	if name[0] == 42 {
		name = name[1:]
	}
	t := &FieldType{
		Type: getType(TName),
		Name: name,
	}
	return t
}

func makeFieldTypeMap(name, key, value string) *FieldType {
	t := &FieldType{
		Type: PTMap,
		Name: name,
		Key:  getType(key),
		Value: &FieldType{
			Type: getType(value),
			Name: value,
		},
	}
	return t
}

func makeFieldTypeSlice(name, value string) *FieldType {
	t := &FieldType{
		Type: PTSlice,
		Name: name,
		Value: &FieldType{
			Type: getType(value),
			Name: value,
		},
	}
	return t
}

func getType(TName string) PType {
	if t, ok := tMap[TName]; ok {
		return t
	}
	return PTStruct
}

func getPBType(name string) (PBType, string, string) {
	s := typeExp.FindStringSubmatch(name)
	if len(s) == 4 {
		mod := s[1]
		subMod := s[3]
		if s[2] == "Tos" {
			return PBTos, mod, subMod
		}
		return PBToc, mod, subMod
	}
	return PBP, name, ""
}

func makeProtoID(modIndex int, mod, subMod string, modMap map[string]int, subModMap map[string]int) (int, int) {
	var modID int
	modIndex, modID = checkMod(mod, modIndex, modMap)
	key := mod + "_" + subMod
	subID := checkSubMod(mod, key, subModMap)
	protoID := modID*100 + subID
	return protoID, modIndex
}

func checkMod(mod string, modIndex int, modMap map[string]int) (int, int) {
	if modID, ok := modMap[mod]; ok {
		return modIndex, modID
	}
	modIndex++
	modMap[mod] = modIndex
	return modIndex, modIndex
}

func checkSubMod(mod string, key string, subModMap map[string]int) int {
	if subModID, ok := subModMap[key]; ok {
		return subModID
	}
	if id, ok := subModMap[mod]; ok {
		id++
		if id > 99 {
			panic(fmt.Errorf("[%s] sub mod lager than 99", mod))
		}
		subModMap[key] = id
		subModMap[mod] = id
		return id
	}
	id := 1
	subModMap[mod] = id
	subModMap[key] = id
	return id
}
