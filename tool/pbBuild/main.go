package main

import (
	"flag"
	"fmt"
	"game/tool/pbBuild/parser"
	"io/ioutil"
	"os"
)

func main() {
	inFilePtr := flag.String("f","","请输入pb_def.go文件路径")
	outDirPtr := flag.String("o","","请输入proto输出路径")
	clientDirPtr := flag.String("c","","请输入client输出路径")
	flag.Parse()
	inFile := *inFilePtr
	outDir := *outDirPtr
	clientDir := *clientDirPtr
	fmt.Println("开始分析协议文件：", inFile)
	f, err := os.Open(inFile)
	if err != nil {
		panic(err)
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	all := parser.Scan(buf)
	if outDir != "" {
		toServer(all,outDir)
	}
	// TODO 导出客户端代码
	if clientDir != "" {
		toLuaClient(all,clientDir)
	}
}


type idString struct {
	id  int
	str string
}

func WriteFile(file,outStr string)  {
	ioFile, err := os.Create(file)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	ioFile.WriteString(outStr)
	fmt.Println("生成目标文件：", file)
}
