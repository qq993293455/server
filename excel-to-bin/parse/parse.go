package parse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"coin-server/common/utils"

	"github.com/xuri/excelize/v2"
)

var schemeJson = map[string][]string{}

func ReadJsonConfig(path string) {
	data, err := ioutil.ReadFile(path)
	utils.Must(err)
	err = json.Unmarshal(data, &schemeJson)
	utils.Must(err)
}

func GetAllExcelName(excelDir string) []string {
	var ps []string
	for k := range schemeJson {
		if k == "drop_mini.xlsx" || k == "drop_lists_mini.xlsx" || k == "roguelike_dungeon_room.xlsx" {
			continue
		}
		ps = append(ps, filepath.Join(excelDir, k))
	}
	return ps
}

type ExcelField struct {
	Name string
	Type string
}

func GetAllExcelField(excelDir string) map[string]*TableInfo {
	out := map[string]*TableInfo{}
	excels := GetAllExcelName(excelDir)
	chOut := make(chan []*TableInfo, len(excels))
	for _, v := range excels {
		go ReadExcel(v, chOut)
	}
	for i := 0; i < len(excels); i++ {
		tis := <-chOut

		for _, v1 := range tis {
			if vv, ok := schemeJson[utils.ToSnake(v1.Name)+".xlsx"]; ok && len(vv) > 0 {
				newFields := make([]*Field, 0, len(v1.Fields))
				tempMap := make(map[string]struct{})
				for _, vv1 := range vv {
					tempMap[vv1] = struct{}{}
				}
				for _, vv2 := range v1.Fields {
					if _, ok := tempMap[vv2.Name]; ok {
						newFields = append(newFields, vv2)
					}
				}
				v1.Fields = newFields
			}
			out[v1.Name] = v1
		}
	}
	return out
}

type TableInfo struct {
	File         string
	Name         string
	Parent       string
	Child        string
	Start        int
	End          int
	Fields       []*Field
	FirstValue   interface{}
	Rows         []interface{}
	DataByte     []byte
	ComplexField bool
}

func NewTableInfo(file, name, parent string) *TableInfo {
	return &TableInfo{
		File:   file,
		Name:   name,
		Parent: parent,
		Fields: make([]*Field, 0),
		Rows:   make([]interface{}, 0),
	}
}

func (t *TableInfo) FieldExist(name string) bool {
	for _, f := range t.Fields {
		if f.Name == name {
			return true
		}
	}
	return false
}

func (t *TableInfo) GetField(index int) (*Field, bool) {
	for _, f := range t.Fields {
		if f.Index == index {
			return f, true
		}
	}
	return nil, false
}

func (t *TableInfo) AddField(field *Field) {
	t.Fields = append(t.Fields, field)
	if !t.ComplexField && field.ComplexField {
		t.ComplexField = true
	}
}

func (t *TableInfo) SetFirstValue(value interface{}) {
	t.FirstValue = value
}

func (t *TableInfo) AddRow(row interface{}) bool {
	t.Rows = append(t.Rows, row)

	return true
}

func (t *TableInfo) Marshal() error {
	var data interface{}
	data = t.Rows
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	t.DataByte = b
	return nil
}

type Field struct {
	Name         string
	Type         string
	Index        int
	Comment      string
	ComplexField bool
}

func Panic(msg string, path string) {
	panic(msg + ":" + path)
}

func ReadExcel(file string, out chan []*TableInfo) {
	f, err := excelize.OpenFile(file)
	if err != nil {
		panic(err)
	}
	tables := make(map[string]*TableInfo)
	tableInfo := make([]*TableInfo, 0)
	sheet := f.GetSheetName(0)
	rootName, err := f.GetCellValue(sheet, "A1")
	if err != nil {
		panic(err)
	}
	if rootName == "" {
		Panic("表名不能为空", file)
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		panic(err)
	}
	if len(rows) < 4 {
		Panic(fmt.Sprintf("表 %s 配置有误", rootName), file)
	}
	var parent string
	tableSort := make([]string, 0)
	for _, name := range rows[0] {
		if name == "" {
			continue
		}
		if _, ok := tables[name]; ok {
			Panic(fmt.Sprintf("重复的表名 %s ", name), file)
		}
		tables[name] = NewTableInfo(file, name, parent)
		tableSort = append(tableSort, name)
		parent = name
	}
	for _, name := range tableSort {
		table := tables[name]
		if table.Parent != "" {
			if _, ok := tables[table.Parent]; !ok {
				Panic(fmt.Sprintf("表 %s 的父表 %s 不存在", table.Name, table.Parent), file)
			}
			tables[table.Parent].Child = table.Name
		}
		tableInfo = append(tableInfo, table)
	}

	if rootName == "key_value" {
		tableInfo = []*TableInfo{parseKeyValue(file, rootName, rows)}
	} else {
		parseNormal(file, tableInfo, rows)
	}
	out <- tableInfo
}

func isCommentColumn(field string) bool {
	return strings.Index(field, "#") > -1
}

func getFieldComment(i int, list []string) string {
	if i >= len(list) {
		return ""
	}
	return list[i]
}

func isComplexField(field string) bool {
	v := !(field == "" || field == "key" || field == "string" || field == "int64" || field == "bool" || strings.Contains(field, "double") || !strings.Contains(field, "repeated"))

	return v
}

func parseKeyValue(file, tableName string, rows [][]string) *TableInfo {
	comments := rows[1]
	types := rows[2]
	fields := rows[3]
	table := NewTableInfo(file, tableName, "")
	for i, field := range fields {
		if isCommentColumn(field) {
			continue
		}
		field = utils.ToSnake(field)
		fields[i] = field
		table.AddField(&Field{
			Name:    field,
			Type:    types[i],
			Index:   0,
			Comment: getFieldComment(i, comments),
		})
	}
	table.AddField(&Field{
		Name:    "comment",
		Type:    "string",
		Index:   0,
		Comment: getFieldComment(1, comments),
	})
	return table
}

func getTableInfo(i int, list []*TableInfo) (*TableInfo, bool) {
	for _, info := range list {
		if i >= info.Start && i <= info.End {
			return info, true
		}
	}
	return nil, false
}

func getTableInfoByName(name string, list []*TableInfo) (*TableInfo, bool) {
	for _, info := range list {
		if info.Name == name {
			return info, true
		}
	}
	return nil, false
}

func getParentPK(name string) string {
	return name + "_id"
}

func parseNormal(file string, tableInfo []*TableInfo, rows [][]string) {
	comments := rows[1]
	tmpTypes := rows[2]
	fields := rows[3]
	types := make([]string, 0)

	var (
		firstIsKey  bool
		hasKeyField bool
	)
	rowLen := 0
	index := make([]int, 0)
	for i := 0; i < len(tmpTypes); i++ {
		v := strings.TrimSpace(tmpTypes[i])
		if strings.Index(v, "#") == -1 {
			rowLen++
		}
		// 注释列，直接将类型设置为string
		// if i < len(fields) && strings.Index(fields[i], "#") > -1 {
		//	v = "string"
		// }

		types = append(types, v)
		if strings.ToLower(v) == "key" {
			index = append(index, i)
			hasKeyField = true
		}
		if i == 0 && strings.ToLower(v) == "key" {
			firstIsKey = true
		}
	}
	if !firstIsKey {
		Panic("第一列字段类型必须为key", file)
	}

	if !hasKeyField {
		Panic("同一张表至少包含一列类型为key", file)
	}
	for i := 0; i < len(index); i++ {
		v := index[i]
		next := len(types) - 1
		if i < len(index)-1 {
			next = index[i+1] - 1
		}
		item := tableInfo[i]
		item.Start = v
		item.End = next
		tableInfo[i] = item
	}
	for i, field := range fields {
		if isCommentColumn(field) || isCommentColumn(tmpTypes[i]) {
			continue
		}

		info, ok := getTableInfo(i, tableInfo)
		if !ok {
			continue
		}
		if field == "" || strings.TrimSpace(field) == "" {
			Panic(fmt.Sprintf("字段名不能为空(%s)", info.Name), info.File)
		}

		if fields[info.Start] != "id" && fields[info.Start] != "Id" && fields[info.Start] != "ID" {
			Panic(fmt.Sprintf("%s 表 key 列对应（第一列）的字段名必须为 id，当前为：%s", info.Name, fields[info.Start]), info.File)
		}
		if info.FieldExist(field) {
			Panic(fmt.Sprintf("%s 表 字段 %s 重复", info.Name, field), info.File)
		}
		if info.Parent != "" {
			parent, ok := getTableInfoByName(info.Parent, tableInfo)
			if ok {
				pkName := getParentPK(parent.Name)
				if field == pkName {
					Panic(fmt.Sprintf("字段名 %s 有误，不能以父表名加 _id 命名", field), info.File)
				}
				if !info.FieldExist(pkName) {
					info.AddField(&Field{
						Name:    pkName,
						Type:    "int64",
						Index:   -1,
						Comment: parent.Name,
					})
				}
			}
		}
		info.AddField(&Field{
			Name:         field,
			Type:         types[i],
			Index:        i,
			Comment:      getFieldComment(i, comments),
			ComplexField: isComplexField(types[i]),
		})
	}
}
