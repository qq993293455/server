package cpb

import (
	"sort"

	"coin-server/common/utils"
	"coin-server/excel-to-bin/parse"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc/builder"
)

var globalDoubleRepeatedInt64 = createDoubleRepeatedInt64()

func createDoubleRepeatedInt64() *builder.MessageBuilder {
	b := builder.NewMessage("DoubleRepeatedInt64")
	f := builder.NewField("slice", builder.FieldTypeInt64()).SetProto3Optional(false).SetRepeated()
	b.AddField(f)
	return b
}

func CreateProtoBuffer(table map[string]*parse.TableInfo, fieldNameFunc func(string) string) *builder.FileBuilder {
	fb := builder.NewFile("out_cpp.proto")
	fb.IsProto3 = true
	fb.Package = "configcpp"
	fb.AddMessage(globalDoubleRepeatedInt64)

	//p := protoprint.Printer{
	//	SortElements: false,
	//}

	keys := make([]string, 0, len(table))
	for k := range table {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	msgTables := builder.NewMessage("Tables")

	msgBuildMap := map[string]*builder.MessageBuilder{}
	for _, v := range keys {
		b := CreateOne(table, v, fieldNameFunc)
		msgBuildMap[v] = b

	}

	for _, v := range keys {
		t := table[v]
		b := msgBuildMap[v]
		if t.Child != "" {
			b.AddField(builder.NewMapField(utils.ToSnake(t.Child), builder.FieldTypeInt64(), builder.FieldTypeMessage(msgBuildMap[t.Child])).SetProto3Optional(false))
		}
		if table[v].Parent == "" {
			if b.GetName() == "KeyValue" {
				fb := builder.NewMapField(utils.ToSnake(b.GetName()), builder.FieldTypeString(), builder.FieldTypeMessage(b))
				msgTables.AddField(fb)
			} else {
				fb := builder.NewMapField(utils.ToSnake(b.GetName()), builder.FieldTypeInt64(), builder.FieldTypeMessage(b))
				msgTables.AddField(fb)
			}
		}
		fb.AddMessage(b)
	}
	msgTables.AddField(builder.NewField("bin_data_version", builder.FieldTypeInt64()))

	fb.AddMessage(msgTables)
	//bDesc, err := fb.Build()
	//utils.Must(err)
	//str, err := p.PrintProtoToString(bDesc)
	//utils.Must(err)
	//fmt.Println(str)
	return fb
}

func CreateOne(ts map[string]*parse.TableInfo, name string, fieldNameFunc func(string) string) *builder.MessageBuilder {
	t := ts[name]
	camelName := utils.ToCamel(t.Name)
	b := builder.NewMessage(camelName)

	for _, v := range t.Fields {
		isRepeated := false
		var ft *builder.FieldType
		var ftv *builder.FieldType
		switch v.Type {
		case "int64":
			ft = builder.FieldTypeInt64()
		case "key":
			if camelName == "KeyValue" {
				ft = builder.FieldTypeString()
			} else {
				ft = builder.FieldTypeInt64()
			}
		case "string":
			ft = builder.FieldTypeString()
		case "repeated int64":
			ft = builder.FieldTypeInt64()
			isRepeated = true
		case "repeated string":
			ft = builder.FieldTypeString()
			isRepeated = true
		case "map<int64,string>":
			ft = builder.FieldTypeInt64()
			ftv = builder.FieldTypeString()
		case "map<int64,int64>":
			ft = builder.FieldTypeInt64()
			ftv = builder.FieldTypeInt64()
		case "map<string,int64>":
			ft = builder.FieldTypeString()
			ftv = builder.FieldTypeInt64()
		case "bool":
			ft = builder.FieldTypeBool()
		case "double repeated int64":
			ft = builder.FieldTypeMessage(globalDoubleRepeatedInt64)
			isRepeated = true
		default:
			panic("unknown type:" + v.Type)
		}
		if ftv == nil {
			f := builder.NewField(fieldNameFunc(v.Name), ft).SetProto3Optional(false)
			if isRepeated {
				f.SetLabel(descriptor.FieldDescriptorProto_LABEL_REPEATED)
			}
			b.AddField(f)
		} else {
			f := builder.NewMapField(fieldNameFunc(v.Name), ft, ftv).SetProto3Optional(false)
			b.AddField(f)
		}
	}
	return b
}
