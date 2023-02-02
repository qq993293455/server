package xml

import (
	"fmt"
	"io/ioutil"
	"reflect"

	"coin-server/common/statistical2/models"
)

const (
	Header = `<?xml version="1.0" encoding="UTF-8"?>
<metalib>
%s
</metalib>`

	Table = `    <table name="%s" comment="%s">
%s
    </table>`

	Field = `        <field name="%s" comment="%s" />`

	FieldWithDic = `        <field name="%s" comment="%s">
%s
        </field>`

	Dic = `            <dic>%s</dic>`

	DicWithField = `            <dic case_field="%s" when_value="%s">%s</dic>`
)

func genTablesXML(models map[string]models.Model) {
	var xml string
	var temp string
	i := 0
	for _, model := range models {
		temp += genModelXML(model, model.Topic(), model.Desc())
		i++
		if i != len(models) {
			temp += "\n"
		}
	}
	xml = fmt.Sprintf(Header, temp)
	if err := ioutil.WriteFile("tables.xml", []byte(xml), 0666); err != nil {
		panic(err)
	}
}

func genModelXML(model models.Model, table, desc string) string {
	var xmlField string
	fieldNum := reflect.TypeOf(model).Elem().NumField()
	for i := 0; i < fieldNum; i++ {
		f := reflect.TypeOf(model).Elem().Field(i)
		xmlField += fmt.Sprintf(Field, f.Tag.Get("json"), f.Tag.Get("desc"))
		if i != fieldNum-1 {
			xmlField += "\n"
		}
	}
	xmlTable := fmt.Sprintf(Table, table, desc, xmlField)
	return xmlTable
}
