package main

import (
	"fmt"
	"io/ioutil"

	"github.com/xuri/excelize/v2"
)

func main() {
	m1 := readErrorCodeFile()
	m2 := readErrorTextFile()
	var notInTextFile string
	for code, info := range m1 {
		if _, ok := m2[code]; !ok {
			notInTextFile += info.numberCode + "\t" + code + "\t" + info.Text + "\n"
		}
	}
	var needDelete string
	for code := range m2 {
		if _, ok := m1[code]; !ok {
			needDelete += code + "\n"
		}
	}
	out := "需要添加至error_text表的error code：\n\n" + notInTextFile + "\n\n"
	out += "需要从error_text表删除的error code：\n\n" + needDelete
	if err := ioutil.WriteFile(fmt.Sprintf("./out.txt"), []byte(out), 0666); err != nil {
		panic(err)
	}
	fmt.Println(out)
}

type codeInfo struct {
	numberCode string
	Text       string
}

func readErrorCodeFile() map[string]*codeInfo {
	f, err := excelize.OpenFile("./xlsx/error_code.xlsx")
	if err != nil {
		panic(err)
	}
	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		panic(err)
	}
	codeMap := make(map[string]*codeInfo)
	for i := 4; i < len(rows); i++ {
		item := rows[i]
		numberCode, code, text := item[0], item[1], item[2]
		if _, ok := codeMap[code]; ok {
			panic(fmt.Errorf("存在重复的error_code：%s 请联系后端程序处理", code))
		}
		codeMap[code] = &codeInfo{
			numberCode: numberCode,
			Text:       text,
		}
	}
	return codeMap
}

func readErrorTextFile() map[string]string {
	f, err := excelize.OpenFile("./xlsx/error_text.xlsx")
	if err != nil {
		panic(err)
	}
	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		panic(err)
	}
	codeMap := make(map[string]string)
	for i := 4; i < len(rows); i++ {
		item := rows[i]
		code, text := item[0], item[1]
		if _, ok := codeMap[code]; ok {
			panic(fmt.Errorf("error_text存在重复的error_code：%s", code))
		}
		codeMap[code] = text
	}
	return codeMap
}
