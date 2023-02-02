package main

import (
	"flag"
	"fmt"

	"coin-server/script/gen-proto-error/parser"
)

//var parseDir = flag.String("parseDir", "proto/service", "解析的proto目录")
//var excelPath = flag.String("excelPath", "excel/xlsx/error_code.xlsx", "生成的excel输出位置")
//var goPath = flag.String("goPath", "../coin-server/common/proto/service/error_code.go", "生成的go代码输出位置")

var parseDir = flag.String("parseDir", "./proto/service", "解析的proto目录")
var excelPath = flag.String("excelPath", "../excel/xlsx/error_code.xlsx", "生成的excel输出位置")
var goPath = flag.String("goPath", "../../coin-server/common/proto/service/error_code.go", "生成的go代码输出位置")

func main() {
	flag.Parse()
	p, err := parser.NewParser(*parseDir)
	if err != nil {
		panic(err)
	}
	err = p.Parse()
	if err != nil {
		panic(err)
	}
	p.Check()
	p.OutputErrorCodeExcel(*excelPath)
	p.OutputErrorCodeGoCode(*goPath)
	fmt.Println(1)
}
