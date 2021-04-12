package main

import (
	"fmt"
	"game/tool/pbBuild/parser"
	"regexp"
	"strings"
)

func toServer(all []*parser.Proto,outDir string)  {
	outBuf, rt := makeAuto(all)
	WriteFile(outDir+"/pb_auto.go",outBuf)

	outBuf = makeRouter(rt)
	WriteFile(outDir+"/router/router.go",outBuf)
}

func makeAuto(all []*parser.Proto) (string, []*idString) {
	routerExp := regexp.MustCompile(`[\t ]*router[\t ]*:[\t ]*([\w.]+)`)

	var tos []*idString
	var toc []*idString
	var rt []*idString
	for _, f := range all {
		s := routerExp.FindStringSubmatch(f.LineC)
		var router string
		if len(s) == 2 {
			router = s[1]
		}
		name := f.Name

		if f.Type == parser.PBTos {
			id := f.ProtoID
			tos = append(tos, &idString{id: id, str: name})
			if router != "" {
				rt = append(rt, &idString{id: id, str: router})
			}
		} else if f.Type == parser.PBToc {
			id := f.ProtoID
			toc = append(toc, &idString{id: id, str: name})
		}
	}
	tosSlice := []string{
		"var TOS = map[int]interface{}{",
	}
	for _, v := range tos {
		tosSlice = append(tosSlice, fmt.Sprintf("\t%d:&global.%s{},", v.id, v.str))
	}
	tosSlice = append(tosSlice, "}")

	tocSlice := []string{
		"var TOC = map[int]interface{}{",
	}
	for _, v := range toc {
		tocSlice = append(tocSlice, fmt.Sprintf("\t%d:&global.%s{},", v.id, v.str))
	}
	tocSlice = append(tocSlice, "}")
	code := "package proto\n\nimport \"game/global\"\n\n// TODO 自动生成，请勿手工修改\n\n" + strings.Join(tosSlice, "\n") + "\n\n" + strings.Join(tocSlice, "\n")
	return code, rt
}

func makeRouter(rt []*idString) string {
	strs := []string{
		"package router\n",
		"import \"game/global\"",
		"import \"game/player\"\n",
		"func MakeRouter()map[int]*global.HandleFunc{",
		"\tvar rs = map[int]*global.HandleFunc{",
	}
	for _, v := range rt {
		s := fmt.Sprintf("\t\t%d:&player.%s,", v.id, v.str)
		strs = append(strs, s)
	}
	strs = append(strs, "\t}\n\treturn rs\n}")
	return strings.Join(strs, "\n")
}