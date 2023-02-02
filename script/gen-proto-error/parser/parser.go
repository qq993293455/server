package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/emicklei/proto"
	"github.com/xuri/excelize/v2"
)

func Walk(srcPath string) ([]string, error) {
	var protos []string
	err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info == nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".proto") {
			protos = append(protos, path)
			return nil
		}
		return err
	})
	return protos, err
}

type Parser struct {
	readers     []*proto.Parser
	messages    map[string]*MessageBase
	enums       map[string]*ErrorInfo
	ext         map[string]*proto.Message
	enumsModels map[string][]*ErrorInfo
	Files       []*File
}

func NewParser(dir string) (*Parser, error) {
	dirs := strings.Split(dir, ",")
	var files []string
	for _, v := range dirs {
		fs, err := Walk(v)
		if err != nil {
			return nil, err
		}
		files = append(files, fs...)
	}

	readers := make([]*proto.Parser, 0, len(files))
	for _, v := range files {
		f, err := os.Open(v)
		if err != nil {
			return nil, err
		}
		p := proto.NewParser(f)
		_, fn := filepath.Split(v)
		p.Filename(fn)
		readers = append(readers, p)
	}
	return &Parser{
		readers:     readers,
		messages:    map[string]*MessageBase{},
		enums:       map[string]*ErrorInfo{},
		ext:         map[string]*proto.Message{},
		enumsModels: map[string][]*ErrorInfo{},
	}, nil
}

func (this_ *Parser) Parse() error {
	for _, v := range this_.readers {
		definition, err := v.Parse()
		if err != nil {
			return err
		}
		fmt.Println(definition.Filename)
		file := &File{FileName: definition.Filename}
		definition.Accept(file)
		this_.Files = append(this_.Files, file)
	}
	return nil
}

func (this_ *Parser) Check() {
	commonErrorMap := map[string]bool{}
	checkDump := map[string]*ErrorInfo{}
	for _, file := range this_.Files {
		for _, mb := range file.Mbs {
			if oldMb, ok := this_.messages[mb.Msg.Name]; ok {
				Errorf("message '%s'[%s:%d] had defined [%s:%d]", mb.Msg.Name, file.FileName, mb.Msg.Position.Line, oldMb.FileName, oldMb.Msg.Position.Line)
			}
			this_.messages[mb.Msg.Name] = mb
		}
		for _, ei := range file.Eis {
			if oldEi, ok := this_.enums[ei.Enum.Name]; ok {
				Errorf("enum value '%s'[%s:%d] had defined [%s:%d]", ei.Enum.Name, file.FileName, ei.Enum.Position.Line, oldEi.FileName, oldEi.Enum.Position.Line)
			}
			this_.enums[ei.Enum.Name] = ei
			if oldEi, ok := checkDump[ei.TipName]; ok {
				Errorf("enum value '%s'[%s:%d] had defined [%s:%d]", ei.TipName, file.FileName, ei.Enum.Position.Line, oldEi.FileName, oldEi.Enum.Position.Line)
			}
			checkDump[ei.TipName] = ei
			this_.enumsModels[ei.Module] = append(this_.enumsModels[ei.Module], ei)
			if ei.Enum.Parent.(*proto.Enum).Name == "CommonErrorCode" {
				commonErrorMap[ei.Enum.Name] = true
			}
		}
		for k, v := range file.Ext {
			this_.ext[k] = v
		}
	}

	for _, v := range this_.Files {
		v.Check(commonErrorMap)
	}

	for k := range this_.enumsModels {
		em := this_.enumsModels[k]
		sort.Slice(em, func(i, j int) bool {
			return em[i].TipName < em[j].TipName
		})
	}
}

func Errorf(format string, a ...interface{}) {
	fmt.Fprintln(os.Stderr, "Parse Failed:", fmt.Sprintf(format, a...))
	os.Exit(-1)
}

func MustNil(err error) {
	if err != nil {
		panic(err)
	}
}

func (this_ *Parser) OutputErrorCodeExcel(outPath string) {
	excel := excelize.NewFile()
	defer func() {
		err := excel.SaveAs(outPath)
		if err != nil {
			Errorf("保存[%s]失败:%v", outPath, err)
		}
	}()
	style, err := excel.NewStyle(`{"alignment":{"horizontal":"left","ident":1,"justify_last_line":true,"reading_order":0,"relative_indent":1,"shrink_to_fit":true,"text_rotation":0,"vertical":"","wrap_text":true},"fill":{"type":"gradient","color":["#C6E0B4","#C6E0B4"],"shading":1}}`)
	MustNil(err)
	style1, err := excel.NewStyle(`{"alignment":{"horizontal":"left","ident":1,"justify_last_line":true,"reading_order":0,"relative_indent":1,"shrink_to_fit":true,"text_rotation":0,"vertical":"","wrap_text":true},"fill":{"type":"gradient","color":["#006400","#006400"],"shading":1}}`)
	MustNil(err)
	style2, err := excel.NewStyle(`{"alignment":{"horizontal":"left","ident":1,"justify_last_line":true,"reading_order":0,"relative_indent":1,"shrink_to_fit":true,"text_rotation":0,"vertical":"","wrap_text":true}}`)
	MustNil(err)
	errorCodeSheetName := "error_code"
	excel.SetActiveSheet(excel.NewSheet(errorCodeSheetName))
	excel.DeleteSheet("Sheet1")
	MustNil(excel.SetColWidth(errorCodeSheetName, "A", "E", 27))
	MustNil(excel.SetCellValue(errorCodeSheetName, "A1", "error_code"))
	MustNil(excel.SetCellValue(errorCodeSheetName, "A2", "key"))
	MustNil(excel.SetCellValue(errorCodeSheetName, "A3", "key"))
	MustNil(excel.SetCellValue(errorCodeSheetName, "A4", "id"))
	MustNil(excel.SetCellValue(errorCodeSheetName, "B2", "服务端英文描述"))
	MustNil(excel.SetCellValue(errorCodeSheetName, "B3", "string"))
	MustNil(excel.SetCellValue(errorCodeSheetName, "B4", "tip_name"))
	MustNil(excel.SetCellValue(errorCodeSheetName, "C2", "客户端中文描述"))
	MustNil(excel.SetCellValue(errorCodeSheetName, "C3", "string"))
	MustNil(excel.SetCellValue(errorCodeSheetName, "C4", "tip_desc"))
	MustNil(excel.SetCellValue(errorCodeSheetName, "D2", "code码"))
	MustNil(excel.SetCellValue(errorCodeSheetName, "D3", "int64"))
	MustNil(excel.SetCellValue(errorCodeSheetName, "D4", "#code"))
	MustNil(excel.SetCellValue(errorCodeSheetName, "E2", "所属模块"))
	MustNil(excel.SetCellValue(errorCodeSheetName, "E3", "string"))
	MustNil(excel.SetCellValue(errorCodeSheetName, "E4", "#module"))
	//MustNil(excel.SetCellValue(errorCodeSheetName, "A3", "string"))
	//MustNil(excel.SetCellValue(errorCodeSheetName, "B3", "string"))
	//MustNil(excel.SetCellValue(errorCodeSheetName, "C3", "int"))
	//MustNil(excel.SetCellValue(errorCodeSheetName, "D3", "string"))
	//MustNil(excel.SetCellValue(errorCodeSheetName, "A4", "PrimaryKey"))
	//MustNil(excel.SetCellValue(errorCodeSheetName, "B4", "output:c$i18n:true"))
	//MustNil(excel.SetCellValue(errorCodeSheetName, "A5", "服务端英文描述"))
	//MustNil(excel.SetCellValue(errorCodeSheetName, "B5", "客户端中文描述"))
	//MustNil(excel.SetCellValue(errorCodeSheetName, "C5", "code码"))
	//MustNil(excel.SetCellValue(errorCodeSheetName, "D5", "所属模块"))
	//MustNil(excel.SetCellValue(errorCodeSheetName, "E5", "测试"))
	index := 4
	//ecSheet:=excel.Sheet[errorCodeSheetName]
	keys := make([]string, 0, len(this_.enumsModels))
	for k := range this_.enumsModels {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	id := 0
	for i := 0; i < len(keys); i++ {
		vs := this_.enumsModels[keys[i]]
		for _, v := range vs {
			index++
			id++
			MustNil(excel.SetCellValue(errorCodeSheetName, createCellAxis("A", index), id))
			MustNil(excel.SetCellValue(errorCodeSheetName, createCellAxis("B", index), v.TipName))
			MustNil(excel.SetCellValue(errorCodeSheetName, createCellAxis("C", index), v.TipDesc))
			ec, err := strconv.Atoi(v.ErrorCode)
			MustNil(err)
			MustNil(excel.SetCellInt(errorCodeSheetName, createCellAxis("D", index), ec))

			MustNil(excel.SetCellValue(errorCodeSheetName, createCellAxis("E", index), v.Module))
		}
	}
	MustNil(excel.SetCellStyle(errorCodeSheetName, createCellAxis("D", 1), createCellAxis("D", index), style2))
	moduleMap := map[string]int{}
	keys = keys[:0]
	for _, v := range this_.messages {
		keys = append(keys, v.Msg.Name)
	}
	sort.Strings(keys)

	for i := 0; i < len(keys); i++ {
		v := this_.messages[keys[i]]
		sheetName := "#" + v.Module
		moduleIndex, ok := moduleMap[sheetName]
		if !ok {
			excel.NewSheet(sheetName)
			MustNil(excel.SetColWidth(sheetName, "A", "B", 30))
			moduleIndex++
			MustNil(excel.SetCellStyle(sheetName, createCellAxis("A", moduleIndex), createCellAxis("B", moduleIndex), style1))
		}
		moduleIndex++
		MustNil(excel.SetCellValue(sheetName, createCellAxis("A", moduleIndex), v.Msg.Name))

		MustNil(excel.SetCellStyle(sheetName, createCellAxis("A", moduleIndex), createCellAxis("B", moduleIndex), style))
		MustNil(excel.SetCellStr(sheetName, createCellAxis("B", moduleIndex), v.Id))

		for _, v := range v.Errors {
			er, ok := this_.enums[v]
			if ok {
				moduleIndex++
				MustNil(excel.SetCellValue(sheetName, createCellAxis("A", moduleIndex), er.TipName))
				MustNil(excel.SetCellValue(sheetName, createCellAxis("B", moduleIndex), er.TipDesc))
			}
		}
		moduleIndex++
		MustNil(excel.SetCellStyle(sheetName, createCellAxis("A", moduleIndex), createCellAxis("B", moduleIndex), style1))
		moduleMap[sheetName] = moduleIndex
	}
	for k, v := range moduleMap {
		for i := 1; i <= v; i++ {
			MustNil(excel.SetRowHeight(k, i, 20))
		}
	}
}

func createCellAxis(col string, row int) string {
	return col + strconv.Itoa(row)
}

func (this_ *Parser) OutputErrorCodeGoCode(outPath string) {
	m := map[string][]*ErrorInfo{}
	for _, v := range this_.enums {
		n := v.Enum.Parent.(*proto.Enum).Name
		m[n] = append(m[n], v)
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
		ei := m[k]
		sort.Slice(ei, func(i, j int) bool {
			return ei[i].Name < ei[j].Name
		})
	}
	sort.Strings(keys)
	//var errors string
	var stackError string
	for i := 0; i < len(keys); i++ {
		//		errorsFunc += fmt.Sprintf(`
		//func (x %s) Error() errmgr.MessageCodeError {
		//	return errorCodeMap[x.String()]
		//}`, keys[i])
		for _, v := range m[keys[i]] {
			//		initErrors += fmt.Sprintf(`
			//errorCodeMap["%s"] = %s`, v.Enum.Name, v.Enum.Name)
			//lowerName := HeadToLower(v.Enum.Name)
			//		errors += fmt.Sprintf(`
			//%s = errmgr.NewError(%s,"%s")`, lowerName, v.ErrorCode, v.TipName)
			stackError += fmt.Sprintf(`
func New%s() *ErrMsg {
	e := &ErrMsg{}
	e.ErrCode = %s
	e.ErrMsg = "%s"
	if openStack {
		e.WithStack()
	}
	return e
}

`, v.Enum.Name, v.ErrorCode, v.TipName)
		}
	}

	str := fmt.Sprintf(`// Code generated by protoc-gen-go. DO NOT EDIT.
// source: %s
package errmsg

import (
	"coin-server/common/values/env"
)

var openStack bool

func init() {
	if env.GetString(env.ERROR_CODE_STACK) == "1" {
		openStack=true
	}
}

%s
`, outPath, stackError)
	err := ioutil.WriteFile(outPath, []byte(str), os.ModePerm)
	if err != nil {
		Errorf("保存[%s]失败:%v", outPath, err)
	}
}

func HeadToLower(s string) string {
	if s[0] >= 'A' && s[0] <= 'Z' {
		x := &strings.Builder{}
		x.WriteByte(s[0] + 32)
		x.WriteString(s[1:])
		return x.String()
	}
	return s
}
