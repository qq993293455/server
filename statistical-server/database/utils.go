package database

import (
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"

	"coin-server/common/statistical2/models"
)

var LastCreate = map[string]int64{}

type buffer struct {
	r         []byte
	runeBytes [utf8.UTFMax]byte
}

func (b *buffer) write(r rune) {
	if r < utf8.RuneSelf {
		b.r = append(b.r, byte(r))
		return
	}
	n := utf8.EncodeRune(b.runeBytes[0:], r)
	b.r = append(b.r, b.runeBytes[0:n]...)
}

func (b *buffer) indent() {
	if len(b.r) > 0 {
		b.r = append(b.r, '_')
	}
}

func underscore(s string) string {
	b := buffer{
		r: make([]byte, 0, len(s)),
	}
	var m rune
	var w bool
	for _, ch := range s {
		if unicode.IsUpper(ch) {
			if m != 0 {
				if !w {
					b.indent()
					w = true
				}
				b.write(m)
			}
			m = unicode.ToLower(ch)
		} else {
			if m != 0 {
				b.indent()
				b.write(m)
				m = 0
				w = false
			}
			b.write(ch)
		}
	}
	if m != 0 {
		if !w {
			b.indent()
		}
		b.write(m)
	}
	return string(b.r)
}

type SQLWithOutDate struct {
	Table string
	Field string
	Value string
	Args  string
}

type SQL struct {
	Stmt string
	Args string
}

func genSQL(model interface{}) SQL {
	var field string
	var value string
	var table string
	if reflect.TypeOf(model).Kind() != reflect.Ptr {
		panic("model must be a pointer")
	}
	table = reflect.TypeOf(model).Elem().Name()
	if table == "" {
		panic(fmt.Errorf("invalid table name"))
	}
	table = underscore(table)

	elem := reflect.TypeOf(model).Elem()
	for i := 0; i < elem.NumField(); i++ {
		v, ok := elem.Field(i).Tag.Lookup("json")
		if !ok {
			panic(fmt.Errorf("invalid tag"))
		}
		field += fmt.Sprintf("`%s`,", v)
		value += `?,`
	}
	field = field[:len(field)-1]
	value = value[:len(value)-1]
	args := fmt.Sprintf("(%s)", value)

	return SQL{
		Stmt: fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)", table, field, value),
		Args: args,
	}
}

func genSQLWithOutDate(model models.Model) SQLWithOutDate {
	var field string
	var value string
	var table string
	if reflect.TypeOf(model).Kind() != reflect.Ptr {
		panic("model must be a pointer")
	}
	table = reflect.TypeOf(model).Elem().Name()
	if table == "" {
		panic(fmt.Errorf("invalid table name"))
	}
	table = underscore(table)

	elem := reflect.TypeOf(model).Elem()
	if elem.NumField()-1 != len(model.ToArgs()) {
		panic(fmt.Errorf("model: <%s> NumField != len(model.ToArgs), %d != %d",
			reflect.TypeOf(model).Elem().Name(), elem.NumField()-1, len(model.ToArgs())))
	}

	for i := 0; i < elem.NumField(); i++ {
		v, ok := elem.Field(i).Tag.Lookup("json")
		if !ok {
			panic(fmt.Errorf("invalid tag"))
		}
		if v == "id" {
			continue
		}
		field += fmt.Sprintf("`%s`,", v)
		value += `?,`
	}
	field = field[:len(field)-1]
	value = value[:len(value)-1]
	args := fmt.Sprintf("(%s)", value)

	return SQLWithOutDate{
		Table: table,
		Field: field,
		Value: value,
		Args:  args,
	}
}

func genStmt(sql SQLWithOutDate, table string, num int) string {
	temp := fmt.Sprintf("INSERT IGNORE INTO %s (%s) VALUES(%s)", table, sql.Field, sql.Value)
	for i := 1; i < num; i++ {
		temp += fmt.Sprintf(",(%s)", sql.Value)
	}
	return temp
}
