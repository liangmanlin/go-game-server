package toGo

import (
	"game/tool/exporter/excel"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"unsafe"
)

var reg = regexp.MustCompile(`[ \t\n\r]`)
var regChild = regexp.MustCompile(`\[\]\*?(\w+)|\*?(\w+$)|map\[\w+\]\*?(\w+)`)
var regField = regexp.MustCompile(`(\w+)[ \t]+(int|string|\[]\*?\w+|map\[.*]\*?\w+|\*\w+)`)

var defStr string

func Export(exc *excel.Excel, outDir string, defFile string) {
	file,err := os.OpenFile(defFile,os.O_RDWR,0666)
	if err != nil {
		panic(err)
	}
	buf,_ := ioutil.ReadAll(file)
	file.Close()
	defStr = string(buf)
	// 先处理结构体数据
	structMap := make(map[string]*excel.Child)
	var typeDef []string
	var keyType,keyName string
	for _, head := range exc.Header {
		if head.Rule == excel.RuleCommon || head.Rule == excel.RuleServer {
			if keyName == "" {
				keyName = head.Name
			}
			if head.Type == excel.FTypeTerm {
				c := readChild(head.TypeName, structMap)
				head.Child = c
			}
			tn := head.TypeName
			if head.Type == excel.FTypeKey{
				tn = "string"
			} else if head.Type == excel.FTypeInt {
				tn = "int32"
			}
			typeDef = append(typeDef, fmt.Sprintf("\t%s %s", excel.FirstUP(head.Name), tn))
			if keyType == ""{
				keyType = head.TypeName
				if head.Type == excel.FTypeInt {
					keyType = "int32"
				}
			}
		}
	}
	if len(typeDef) == 0 {
		return
	}
	var exportAll []string
	for _, row := range exc.AllData {
		var data []string
		for i, val := range row {
			head := exc.Header[i]
			if head.Rule != excel.RuleServer && head.Rule != excel.RuleCommon {
				continue
			}
			if head.Type == excel.FTypeKey{
				val = excel.CheckValString(val)
			}else if head.Child != nil {
				// 需要展开这个数据
				val = explainChild(val, head.Child)
			}
			data = append(data, fmt.Sprintf("\"%s\":%s", head.Name, val))
		}
		exportAll = append(exportAll, fmt.Sprintf("{%s}", strings.Join(data, ",")))
	}
	name := excel.FirstUP(exc.Name)
	jsonStr := fmt.Sprintf("[\n%s\n]", strings.Join(exportAll, ",\n"))
	file2, err := os.Create(outDir + "/json/" + name + ".json")
	if err != nil {
		panic(err)
	}
	fmt.Printf("导出配置文件：%s.json\n",name)
	defer file2.Close()
	file2.WriteString(jsonStr)
	var getCode,loadCode,keyType2 string
	keyType2 = keyType
	if keyType == "key"{
		keyType2 = "string"
	}
	configName := firstDown(name)
	loadCode = fmt.Sprintf(`func (%s %sConfig)load(path string)  {
	c := &%sConfig{m:make(map[%s]int)}
	if err:= json.Unmarshal(readFile("%s",path), &c.arr);err != nil{panic(err)}
	for i:=0;i<len(c.arr);i++{
		c.m[c.arr[i].%s] = i
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&%s)),*(*unsafe.Pointer)(unsafe.Pointer(&c)))
}`,name[0:1],configName,configName,keyType2,name,keyName,name)
	if keyType != "key"{
		getCode = fmt.Sprintf("type %sConfig struct{\n" +
			"\tm map[%s]int\n" +
			"\tarr []Def%s\n"+
			"}\n"+
			"var %s = &%sConfig{}\n" +
			"func (%s %sConfig)Get(key %s)*Def%s{\n" +
			"\treturn &%s.arr[%s.m[key]]\n}\n%s\n",configName,keyType,name,name,configName,name[0:1],configName,keyType,name,name[0:1],name[0:1],loadCode)
	}else{
		getCode = fmt.Sprintf("type %sConfig struct{\n" +
			"\tm map[string]int\n" +
			"\tarr []Def%s\n"+
			"}\n"+
			"var %s = &%sConfig{}\n" +
			"func (%s %sConfig)Get(key...interface{})*Def%s{\n" +
			"\treturn &%s.arr[%s.m[sliceToString(key)]]\n}\n%s\n",configName,name,name,configName,name[0:1],configName,name,name[0:1],name[0:1],loadCode)
	}
	getCode+= fmt.Sprintf(`func (%s %sConfig)All()[]Def%s {
	return %s.arr
}`,name[0:1],configName,name,name[0:1])
	def := fmt.Sprintf("//%s-start\ntype Def%s struct {\n%s\n}\n%s\n//%s-end",
		name, name, strings.Join(typeDef, "\n"),getCode,name)
	// 替换定义文件
	if strings.Index(defStr, fmt.Sprintf("//%s-start", name)) >= 0 {
		defReg := regexp.MustCompile(fmt.Sprintf("//%s-start[\\s\\S]*//%s-end", name, name))
		defStr = defReg.ReplaceAllString(defStr, def)
	} else {
		defStr += def + "\n"
	}
	if strings.Index(defStr, fmt.Sprintf("PtrMap[\"%s\"]", name)) < 0 {
		defStr = strings.Replace(defStr,"//init-ptr-start","//init-ptr-start\n\t"+fmt.Sprintf("PtrMap[\"%s\"] = %s", name,name),1)
	}
	file,_ = os.OpenFile(defFile,os.O_RDWR|os.O_TRUNC,0666)
	file.WriteString(defStr)
}

func explainChild(val string, child *excel.Child) string {
	// 删除所有空白和换行符
	val = reg.ReplaceAllString(val, "")
	switch child.Type {
	case excel.HeadChildTypeSlice:
		return transSlice(val, child)
	case excel.HeadChildTypeMap:
		return transMap(val, child)
	default:
		return transStruct(val, child)
	}
}

func transSlice(val string, child *excel.Child) string {
	lens := len(val)
	if lens == 0 {
		return "[]"
	}
	if child.Value != excel.FTypeTerm {
		return "[" + val[1:lens-1] + "]"
	}
	// 原则上不允许嵌套，所以这里一定是struct
	var idx, i int
	var buf []byte
	valBuf := *(*[]byte)(unsafe.Pointer(&val))
	lens = len(valBuf)
	count := 0
	for i < lens {
		b := valBuf[i]
		i++
		if b == '{' && count == 0 {
			buf = append(buf, '[')
			count++
		} else if b == '{' {
			buf, i, idx = transField(buf, i, child, b, valBuf, 0)
			count++
		} else if b == '}' && count > 1 {
			buf = append(buf, b)
			count--
		} else if b == '}' {
			buf = append(buf, ']')
			break
		} else if b == ',' && count > 1 {
			buf, i, idx = transField(buf, i, child, b, valBuf, idx)
		} else {
			buf = append(buf, b)
		}
	}
	return string(buf)
}

func transMap(val string, child *excel.Child) string {
	lens := len(val)
	if lens == 0 {
		return "{}"
	}
	// 原则上不允许嵌套，所以这里一定是struct
	var idx, i int
	var buf []byte
	valBuf := *(*[]byte)(unsafe.Pointer(&val))
	lens = len(valBuf)
	count := 0
	for i < lens {
		b := valBuf[i]

		i++
		if b == '{' && count == 0 {
			buf = append(buf, b)
			if valBuf[i]!= '}' && valBuf[i] != '"'{
				buf = append(buf, '"')
			}
			count++
		} else if b == '{' {
			count++
			buf, i, idx = transField(buf, i, child, b, valBuf, 0)
		} else if b == '}' && count > 1 {
			buf = append(buf, b)
			count--
		} else if b == '}' {
			buf = append(buf, b)
			break
		} else if b == ',' && count > 1 {
			buf, i, idx = transField(buf, i, child, b, valBuf, idx)
		}else if b == ','{
			buf = append(buf, b)
			if valBuf[i] != '"'{
				buf = append(buf, '"')
			}
		} else if b == ':'{
			l := len(buf)
			if l > 0 && buf[l-1] != '"'{
				buf = append(buf, '"')
			}
			buf = append(buf, b)
		} else {
			buf = append(buf, b)
		}
	}
	return string(buf)
}

func transStruct(val string, child *excel.Child) string {
	if len(val) == 0 {
		return "{}"
	}
	var idx, i int
	var buf []byte
	valBuf := *(*[]byte)(unsafe.Pointer(&val))
	lens := len(valBuf)
	for i < lens {
		b := valBuf[i]
		i++
		if b == '{' {
			buf, i, idx = transField(buf, i, child, b, valBuf, 0)
		} else if b == ',' {
			buf, i, idx = transField(buf, i, child, b, valBuf, idx)
		} else if b == '}' {
			buf = append(buf, b)
			break
		} else {
			buf = append(buf, b)
		}
	}
	return string(buf)
}

func transField(buf []byte, i int, child *excel.Child, b byte, valBuf []byte, idx int) ([]byte, int, int) {
	buf = append(buf, b, '"')
	field := child.Fields[idx]
	idx++
	buf = append(buf, field.Name...)
	buf = append(buf, '"', ':')
	if field.Child != nil {
		var data string
		data, i = transChild(i, valBuf, field.Child)
		buf = append(buf, data...)
	}
	return buf, i, idx
}

func transChild(i int, valBuf []byte, child *excel.Child) (string, int) {
	var val string
	val, i = getData(valBuf, i)
	switch child.Type {
	case excel.HeadChildTypeSlice:
		return transSlice(val, child), i
	case excel.HeadChildTypeMap:
		return transMap(val, child), i
	default:
		return transStruct(val, child), i
	}
}

func getData(valBuf []byte, i int) (string, int) {
	startI := i
	endI := 0
	structCount := 0
	for i < len(valBuf) {
		b := valBuf[i]
		i++
		if b == '{' {
			structCount++
		} else if b == '}' {
			structCount--
			if structCount == 0 {
				endI = i
				break
			}
		}
	}
	return string(valBuf[startI:endI]), i
}

func readChild(typeName string, structMap map[string]*excel.Child) *excel.Child {
	matches := regChild.FindStringSubmatch(typeName)
	lens := len(matches)
	if lens >= 2 && matches[1] != "" {
		child := &excel.Child{Type: excel.HeadChildTypeSlice}
		readFields(matches[1], structMap, child)
		return child
	}
	if lens >= 3 && matches[2] != "" {
		child := &excel.Child{Type: excel.HeadChildTypeStruct}
		readFields(matches[2], structMap, child)
		return child
	}
	if lens >= 4 && matches[3] != "" {
		child := &excel.Child{Type: excel.HeadChildTypeMap}
		readFields(matches[3], structMap, child)
		return child
	}
	return nil
}

func readFields(cv string, structMap map[string]*excel.Child, child *excel.Child) {
	if cv == "int" || cv == "string" {
		child.Value = excel.FTypeInt // 映射为基础类型就可以了
	} else {
		child.Value = excel.FTypeTerm
		// 从定义文件中继续展开
		if c, ok := structMap[cv]; ok {
			child.Fields = c.Fields
		} else {
			child.Fields = searchChildFromDef(structMap, cv)
		}
	}
}

func searchChildFromDef(structMap map[string]*excel.Child, cv string) []*excel.Field {
	r := regexp.MustCompile(fmt.Sprintf("type[ \t]+%s[ \t]+struct[ \t\n]*{([\\[\\]*/\\w\\r\\n\\s]*)}", cv))
	matches := r.FindStringSubmatch(defStr)
	if len(matches) != 2 {
		panic(fmt.Errorf("未能找到结构定义：%s", cv))
	}
	sl := regField.FindAllStringSubmatch(matches[1], -1)
	child := &excel.Child{Type: excel.HeadChildTypeStruct, Value: excel.FTypeTerm}
	structMap[cv] = child
	var fields []*excel.Field
	tmp := make(map[string]*excel.Field)
	for _, v := range sl {
		if len(v) == 3 {
			f := &excel.Field{Name: v[1]}
			if v[2] != "int" && v[2] != "string" {
				tmp[v[2]] = f // 后续处理,规避环
			}
			fields = append(fields, f)
		}
	}
	for k, v := range tmp {
		v.Child = readChild(k, structMap)
	}
	child.Fields = fields
	return fields
}

func firstDown(name string)string {
	b := name[0]
	if b >= 'a' && b <= 'z' {
		return name
	}
	return string(b+('a'-'A'))+name[1:]
}