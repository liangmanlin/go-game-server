package main

import (
	"game/tool/exporter/excel"
	"game/tool/exporter/toGo"
	"flag"
	"fmt"
	"os"
)

func main() {
	inFilePtr := flag.String("f", "", "请输入excel文件")
	outDirPtr := flag.String("os", "", "可选，请输入后端输出路径")
	clientDirPtr := flag.String("oc", "", "可选，请输入client输出路径")
	flag.Parse()
	if *inFilePtr == "" {
		flag.Usage()
		os.Exit(1)
	}
	exc := excel.LoadFile(*inFilePtr)
	fmt.Printf("head len :%d\n", len(exc.Header))
	if *outDirPtr != "" {
		toGo.Export(exc, *outDirPtr, *outDirPtr+"/def.go")
	}
	if *clientDirPtr != "" {

	}
}
