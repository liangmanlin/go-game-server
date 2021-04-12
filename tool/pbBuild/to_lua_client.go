package main

import (
	"fmt"
	"game/tool/pbBuild/parser"
	"os"
	"sort"
	"strconv"
	"strings"
)

func toLuaClient(all []*parser.Proto, outDir string) {
	outDir = outDir + "/Proto"
	_ = os.MkdirAll(outDir, os.ModePerm)
	id2Name := []string{"NameDic={"}
	tocDic := []string{"TocDic={"}
	var tosEncodeCode []string
	for _, v := range parser.TypeNameMap {
		tosEncodeCode = append(tosEncodeCode, fmt.Sprintf("---@class %s %s", v, v))
	}
	sort.Strings(tosEncodeCode)
	tosEncodeCode = append(tosEncodeCode, "", "local eb = {}", "local pack = string.pack")
	tocDecodeCode := []string{"local unpack = string.unpack", ""}
	tosEbID := 1
	enPMap := make(map[string]bool)
	m := make(map[int]bool)
	var pList []*parser.Proto
	// 生成p协议字典
	pMap := make(map[string]*parser.Proto)
	for _, p := range all {
		if !m[p.ProtoID] && (p.Type == parser.PBToc || p.Type == parser.PBTos) {
			m[p.ProtoID] = true
			id2Name = append(id2Name, fmt.Sprintf("[%d]=\"%s%%s%s\",", p.ProtoID, p.Mod, p.SubMod))
		}
		if p.Type == parser.PBToc {
			tocDic = append(tocDic, fmt.Sprintf("[%d]=%s,", p.ProtoID, p.Name))
			tocDecodeCode = append(tocDecodeCode, decodeCode(p))
		}
		// 生成tos 编码
		if p.Type == parser.PBTos {
			tosEncodeCode = append(tosEncodeCode, encodeCode(p, enPMap, tosEbID, true))
			tosEbID++
		}

		if p.Type == parser.PBP {
			pList = append(pList, p)
			pMap[p.Name] = p
		}
	}
	id2Name = append(id2Name, "}\n")
	tocDic = append(tocDic, "}\n")

	WriteFile(outDir + "/NameDic.lua",strings.Join(id2Name, "\n"))

	WriteFile(outDir + "/TocDic.lua",strings.Join(tocDic, "\n"))

	WriteFile(outDir + "/AllTos.lua",strings.Join(tosEncodeCode, "\n"))
	// 生成p协议encode
	// 先递归构造所有encode
	enPMap2 := make(map[string]bool)
	for k, _ := range enPMap {
		p := pMap[k]
		enPMap2[k] = true
		for _, f := range p.Fields {
			if f.FType.IsBase() {
				continue
			}
			if f.FType.IsArray() {
				if f.FType.Value.Type == parser.PTStruct {
					enterMap(f.FType.Value.Name, enPMap, enPMap2)
				}
			} else if f.FType.IsMap() {
				if f.FType.Value.Type == parser.PTStruct {
					enterMap(f.FType.Value.Name, enPMap, enPMap2)
				}
			} else {
				enterMap(f.FType.Name, enPMap, enPMap2)
			}
		}
	}
	pCodes := []string{"local pack = string.pack", "local unpack = string.unpack", ""}
	for _, p := range pList {
		pCodes = append(pCodes, pCode(p, enPMap2))
	}
	WriteFile(outDir + "/AllP.lua",strings.Join(pCodes, "\n"))

	WriteFile(outDir + "/AllToc.lua",strings.Join(tocDecodeCode, "\n"))

	WriteFile(outDir + "/Proto.lua",protoC())
}

func enterMap(name string, enPMap map[string]bool, enPMap2 map[string]bool) {
	_, ok1 := enPMap[name]
	_, ok2 := enPMap2[name]
	if ok1 || ok2 {

	} else {
		enPMap2[name] = true
	}
}

func pCode(p *parser.Proto, enPMap2 map[string]bool) string {
	var code []string
	if enPMap2[p.Name] {
		code = append(code, encodeCode(p, nil, 0, false))
	}
	// 生成decode
	code = append(code, decodeCode(p))
	return strings.Join(code, "\n")
}

func encodeCode(p *parser.Proto, enPMap map[string]bool, tosEbID int, flag bool) string {
	var head1, head2, head3, head4 string
	if flag {
		head1 = fmt.Sprintf("eb[%d] = function (self)", tosEbID)
		head2 = fmt.Sprintf("eb[%d]", tosEbID)
		head3 = formatHeadC(p.HeadC) + "\n" + fmt.Sprintf("%s = function()", p.Name)
		head4 = fmt.Sprintf("\t---@class %s", p.Name)
	} else {
		head1 = fmt.Sprintf("local %s_Encode = function (self)", p.Name)
		head2 = fmt.Sprintf("%s_Encode", p.Name)
		head3 = fmt.Sprintf("%sTos = function()", p.Name)
		head4 = fmt.Sprintf("\t---@type %s", p.Name)
	}
	code := []string{"--################################",
		fmt.Sprintf("---@param self %s", p.Name),
		head1,
	}
	def := []string{
		"",
		head3,
		head4,
	}
	var ens, fields []string
	if flag {
		ens = []string{"H"} // 第一个是协议编号
		fields = []string{strconv.Itoa(p.ProtoID)}
	}
	var isCreateBuf bool
	for _, f := range p.Fields {
		if flag {
			def = append(def, fmt.Sprintf("\t---@field %s %s @%s", f.Name, getFTypeName(f), f.LineC))
		}
		// 如果是基础类型，可以直接编码，并且要求一定要赋值
		if f.FType.IsBase() {
			if f.FType.Type == parser.PTBool {
				code = append(code, fmt.Sprintf("\tlocal %s_bool = 0\n\tif self.%s then\n\t\t%s_bool = 1 \n\tend", f.Name, f.Name, f.Name))
				fields = append(fields, fmt.Sprintf("%s_bool", f.Name))
				ens = append(ens, "B")
			} else {
				fields = append(fields, fmt.Sprintf("self.%s", f.Name))
				ens = append(ens, getEns(f.FType.Type))
			}
		} else {
			// 先封闭代码
			if len(fields) > 0 {
				rs := getResult(isCreateBuf, ens, fields)
				isCreateBuf = true
				fields = nil
				ens = nil
				code = append(code, rs)
			} else if !isCreateBuf {
				isCreateBuf = true
				code = append(code, "\tlocal buf = ''")
			}
			// 数组（slice）
			if f.FType.IsArray() {
				fv := f.FType.Value
				code = append(code, fmt.Sprintf("\tif self.%s then\n\t\tlocal bufIn = ''\n\t\tfor i=1,#self.%s do", f.Name, f.Name))
				if fv.IsBase() {
					if fv.Type == parser.PTBool {
						code = append(code, "local tmp = 0")
						code = append(code, fmt.Sprintf("if self.%s[i] then\n tmp = 1\nend", f.Name))
						code = append(code, "\t\t\t bufIn = bufIn..pack(\">B\",tmp)")
					} else {
						code = append(code, fmt.Sprintf("\t\t\t bufIn = bufIn..pack(\">%s\",self.%s[i])", getEns(fv.Type), f.Name))
					}
				} else {
					// 肯定是struct，约定不能嵌套数组，map
					if flag {
						enPMap[fv.Name] = true
					}
					code = append(code, fmt.Sprintf("\t\t\tbufIn = bufIn..self.%s[i]:encode()", f.Name))
				}
				code = append(code, "\t\tend")
				code = append(code, "\t\tbuf = buf..pack(\">H\",#bufIn)..bufIn")
				code = append(code, "\telse\n\t\tbuf = buf..pack(\">H\",0)\t\n\tend")
			} else if f.FType.Type == parser.PTMap {
				fv := f.FType.Value
				code = append(code, fmt.Sprintf("\tif self.%s then\n\t\tlocal bufIn = ''\n\t\tfor k,v in pairs(self.%s) do", f.Name, f.Name))
				if fv.IsBase() {
					if fv.Type == parser.PTBool {
						code = append(code, "local tmp = 0")
						code = append(code, "if v then\n tmp = 1\nend")
						code = append(code, fmt.Sprintf("\t\t\t bufIn = bufIn..pack(\">%sB\",k,tmp)", getEns(f.FType.Key)))
					} else {
						code = append(code, fmt.Sprintf("\t\t\t bufIn = bufIn..pack(\">%s%s\",k,v)", getEns(f.FType.Key), getEns(fv.Type)))
					}
				} else {
					// 肯定是struct，约定不能嵌套数组，map
					if flag {
						enPMap[fv.Name] = true
					}
					code = append(code, fmt.Sprintf("\t\t\t bufIn = bufIn..pack(\">%s\",k)..v:encode()", getEns(f.FType.Key)))
				}
				code = append(code, "\t\tend")
				code = append(code, "\t\tbuf = buf..pack(\">H\",#bufIn)..bufIn")
				code = append(code, "\telse\n\t\tbuf = buf..pack(\">H\",0)\t\n\tend")
			} else {
				// 肯定是struct，约定不能嵌套数组，map
				if flag {
					enPMap[f.FType.Name] = true
				}
				code = append(code, fmt.Sprintf("\tif self.%s then\n\t\tbuf = buf..pack(\">B\",1)..self.%s:encode()", f.Name, f.Name))
				code = append(code, "\telse\n\t\tbuf = buf..pack(\">B\",0)\t\n\tend")
			}
		}
	}
	if len(ens) > 0 {
		rs := getResult(isCreateBuf, ens, fields)
		code = append(code, rs)
	}
	if flag {
		code = append(code, fmt.Sprintf("\treturn buf,%d",p.ProtoID), "end")
	} else if len(p.Fields) > 0 {
		code = append(code, "\treturn buf", "end")
	} else {
		code = append(code, "\treturn ''", "end")
	}
	def = append(def, "\tlocal f = {}")
	def = append(def, fmt.Sprintf("\tf.encode = %s\n\treturn f\nend", head2))
	return strings.Join(code, "\n") + strings.Join(def, "\n")
}

func decodeCode(p *parser.Proto) string {
	def := []string{
		fmt.Sprintf("\t---@class %s", p.Name),
	}
	var local, result, code, ens, fields []string
	var hasSize bool
	for _, f := range p.Fields {
		def = append(def, fmt.Sprintf("\t---@field %s %s @%s", f.Name, getFTypeName(f), f.LineC))
		local = append(local, f.Name)
		// 如果是基础类型，可以直接编码，并且要求一定要赋值
		if f.FType.IsBase() {
			if f.FType.Type == parser.PTBool {
				result = append(result, fmt.Sprintf("\t\t%s = %s > 0, \t-- %s %s", f.Name, f.Name, parser.TypeNameMap[f.FType.Type], f.LineC))
			} else {
				result = append(result, fmt.Sprintf("\t\t%s = %s, \t-- %s %s", f.Name, f.Name, parser.TypeNameMap[f.FType.Type], f.LineC))
			}
			ens = append(ens, getEns(f.FType.Type))
			fields = append(fields, f.Name)
		} else {
			result = append(result, fmt.Sprintf("\t\t---@type %s %s", getFTypeName(f), f.LineC))
			result = append(result, fmt.Sprintf("\t\t%s = %s,", f.Name, f.Name))
			// 数组（slice）
			if f.FType.IsArray() {
				fv := f.FType.Value
				hasSize, ens, fields, code = buildSize(code, ens, fields, "H", hasSize)
				code = append(code, fmt.Sprintf("\t%s = {}", f.Name))
				code = append(code, "\tif _size > 0 then")
				code = append(code, "\t\tlocal lastPos = b_pos + _size")
				code = append(code, "\t\tlocal i = 1")
				code = append(code, "\t\trepeat")
				if fv.IsBase() {
					code = append(code, fmt.Sprintf("\t\t\t%s[i],b_pos = unpack(\">%s\",buf,b_pos)", f.Name, getEns(fv.Type)))
				} else {
					// 不允许嵌套,所以这里一定是struct
					code = append(code, fmt.Sprintf("\t\t\t%s[i],b_pos = %s(buf,b_pos)", f.Name, fv.Name))
				}
				code = append(code, "\t\t\ti = i + 1")
				code = append(code, "\t\tuntil(b_pos >= lastPos)")
				code = append(code, "\tend")
			} else if f.FType.IsMap() {
				fv := f.FType.Value
				hasSize, ens, fields, code = buildSize(code, ens, fields, "H", hasSize)
				code = append(code, fmt.Sprintf("\t%s = {}", f.Name))
				code = append(code, "\tif _size > 0 then")
				code = append(code, "\t\tlocal lastPos = b_pos + _size")
				code = append(code, "\t\tlocal k,v")
				code = append(code, "\t\trepeat")
				if fv.IsBase() {
					code = append(code, fmt.Sprintf("\t\t\tk,v,b_pos = unpack(\">%s%s\",buf,b_pos)", getEns(f.FType.Key), getEns(fv.Type)))
					code = append(code, fmt.Sprintf("\t\t\t%s[k] = v", f.Name))
				} else {
					// 不允许嵌套,所以这里一定是struct
					code = append(code, fmt.Sprintf("\t\t\tk,b_pos = unpack(\">%s\",buf,b_pos)", getEns(f.FType.Key)))
					code = append(code, fmt.Sprintf("\t\t\tv,b_pos = %s(buf,b_pos)", fv.Name))
					code = append(code, fmt.Sprintf("\t\t\t%s[k] = v", f.Name))
				}
				code = append(code, "\t\tuntil(b_pos >= lastPos)")
				code = append(code, "\tend")
			} else {
				// 这里一定是struct了
				hasSize, ens, fields, code = buildSize(code, ens, fields, "B", hasSize)
				code = append(code, "\tif _size > 0 then")
				code = append(code, fmt.Sprintf("\t\t%s,b_pos = %s(buf,b_pos)", f.Name, f.FType.Name))
				code = append(code, "\tend")
			}
		}
	}
	if len(fields) > 0 {
		hasSize, ens, fields, code = buildSize(code, ens, fields, "", hasSize)
	}
	code = append(code, fmt.Sprintf("\t---@class %s", p.Name))
	code = append(code, "\tlocal f = {")
	code = append(code, result...)
	code = append(code, "\t}\n\treturn f,b_pos\nend")
	codeHead := []string{
		formatHeadC(p.HeadC),
		fmt.Sprintf("%s = function(buf,b_pos)", p.Name),
	}
	if len(local) > 0 {
		codeHead = append(codeHead,"\tlocal " + strings.Join(local, ","))
	}
	codeHead = append(codeHead, code...)
	return strings.Join(codeHead, "\n")
}

func buildSize(code, ens, fields []string, head string, hasSize bool) (bool, []string, []string, []string) {
	if head != "" {
		code = checkHaseSize(code, hasSize)
		ens = append(ens, head)
		fields = append(fields, "_size")
	}
	rs := getDecodeResult(ens, fields)
	code = append(code, rs)
	return true, nil, nil, code
}

func checkHaseSize(code []string, hasSize bool) []string {
	if hasSize {
		return code
	}
	return append(code, "\tlocal _size")
}

func getEns(t parser.PType) string {
	switch t {
	case parser.PTBool:
		return "B"
	case parser.PTInt8:
		return "b"
	case parser.PTInt16:
		return "h"
	case parser.PTInt32:
		return "i4"
	case parser.PTInt64:
		return "i8"
	case parser.PTUInt16:
		return "H"
	case parser.PTFloat32:
		return "f"
	case parser.PTFloat64:
		return "d"
	case parser.PTString:
		return "s2"
	}
	return ""
}

func getResult(isCreateBuf bool, ens, fields []string) string {
	if isCreateBuf {
		return fmt.Sprintf("\tbuf = buf..pack(\">%s\",%s)", strings.Join(ens, ""), strings.Join(fields, ","))
	}
	return fmt.Sprintf("\tlocal buf = pack(\">%s\",%s)", strings.Join(ens, ""), strings.Join(fields, ","))
}

func getFTypeName(f *parser.Field) string {
	if f.FType.IsBase() {
		return f.FType.Name
	}
	if f.FType.Type == parser.PTSlice {
		return f.FType.Value.Name + "[]"
	}
	if f.FType.Type == parser.PTMap {
		if f.FType.Value.Type == parser.PTStruct {
			return fmt.Sprintf("table<%s,%s>", parser.TypeNameMap[f.FType.Key], f.FType.Value.Name)
		}
		return fmt.Sprintf("table<%s,%s>", parser.TypeNameMap[f.FType.Key], f.FType.Value.Name)
	}
	return f.FType.Name
}

func formatHeadC(headC string) string {
	return "--" + strings.ReplaceAll(headC, "\n", "\n--")
}

func getDecodeResult(ens, fields []string) string {
	fields = append(fields, "b_pos")
	return fmt.Sprintf("\t%s = unpack(\">%s\",buf,b_pos)", strings.Join(fields, ","), strings.Join(ens, ""))
}

func protoC() string {
	return "local unpack = string.unpack\n\n" +
		"-- 负责解包\n" +
		"function ProtoDecode(buf)\n" +
		"    local protoID,b_pos = unpack(\">H\",buf,1)\n" +
		"    local decodeFunc = TocDic[protoID]\n" +
		"    if decodeFunc then\n" +
		"        local p = decodeFunc(buf,b_pos)\n" +
		"        return p,protoID\n" +
		"    end\n" +
		"end"
}