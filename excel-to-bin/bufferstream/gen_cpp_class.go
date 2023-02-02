package bufferstream

import (
	"fmt"
	"io/ioutil"
	"math"
	"sort"
	"strings"

	"coin-server/common/utils"
	"coin-server/common/values/env"
	env2 "coin-server/excel-to-bin/env"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
)

type field struct {
	TypeName  string
	FieldName string
}

type class struct {
	TypeName string
	Fields   []field
}

var (
	classes = map[string]class{
		"DoubleRepeatedInt64": {
			TypeName: "DoubleRepeatedInt64",
			Fields: []field{{
				TypeName:  "std::vector<int64_t>",
				FieldName: "slice",
			}},
		},
	}
)

var classOrder = map[string]int{
	"DoubleRepeatedInt64":  1,
	"DropMini":             2,
	"DropListsMini":        3,
	"RoguelikeDungeonRoom": 4,
	"Attr":                 1005,
	"AttrTrans":            1006,
	"BossHall":             1007,
	"Buff":                 1008,
	"BuffEffect":           1009,
	"Drop":                 1010,
	"DropLists":            1011,
	"Dungeon":              1012,
	"InhibitAtk":           1013,
	"KeyValue":             1014,
	"MapScene":             1015,
	"Mechanics":            1016,
	"Medicament":           1017,
	"Monster":              1018,
	"MonsterGroup":         1019,
	"Robot":                1020,
	"RoguelikeArtifact":    1021,
	"RoguelikeDungeon":     1022,
	"RoleLv":               1023,
	"RoleReachDungeon":     1024,
	"RowHero":              1025,
	"Skill":                1026,
	"Summoned":             1027,
	"TempBag":              1028,
	"Tables":               math.MaxInt,
}

func GenCppClass(tableDesc *desc.MessageDescriptor) {
	genClass(tableDesc)
	var str string
	str += "#ifndef CONFIG_STREAM_HPP\n"
	str += "#define CONFIG_STREAM_HPP\n"
	str += "#include <vector>\n"
	str += "#include <unordered_map>\n"
	str += "#include \"stream_base.hpp\"\n"
	str += "namespace streambuffer {\n"
	var ks []string

	for k := range classes {
		ks = append(ks, k)
	}
	sort.Slice(ks, func(i, j int) bool {
		i1, oki := classOrder[ks[i]]
		j1, okj := classOrder[ks[j]]
		if oki && okj {
			return i1 < j1
		}
		if i1 == math.MaxInt {
			return false
		}
		if j1 == math.MaxInt {
			return true
		}
		if oki {
			return true
		}
		if okj {
			return false
		}
		return ks[i] < ks[j]
	})
	header := ""
	for _, k := range ks {
		header += "class " + k + ";\n"
	}

	str += header + "\n\n"

	strImplement := ""

	for _, k := range ks {
		v := classes[k]
		str1 := fmt.Sprintf("class %s:public streambuffer::ReadBaseStreamI, public streambuffer::WriteBaseStreamI {\n", k)
		str1 += fmt.Sprintf("public:\n")
		readStr := "rbs"
		writeStr := "wbs"
		for _, v1 := range v.Fields {
			realField := strings.ReplaceAll(v1.FieldName, "_", "")
			str1 += fmt.Sprintf("\t%s %s;\n", v1.TypeName, realField)
			readStr += ">>" + realField
			writeStr += "<<" + realField
		}
		str1 += fmt.Sprintf("\tvirtual ~%s() {}\n", k)
		str1 += "\tvirtual void ReadStream(streambuffer::RBStream&);\n"
		str1 += "\tvirtual void WriteStream(streambuffer::WBStream&)const;\n"
		strImplement += fmt.Sprintf(`	void %s::ReadStream(streambuffer::RBStream& rbs){
		%s;
	}
`, k, readStr)
		strImplement += fmt.Sprintf(`	void %s::WriteStream(streambuffer::WBStream& wbs)const {
		%s;
	}
`, k, writeStr)

		str1 += "};\n\n"
		str += str1
	}
	str += strImplement
	str += "}\n"
	str += "#endif \n\n"
	utils.Must(ioutil.WriteFile(env.GetString(env2.CPP_CONFIG_CODE_PATH), []byte(str), 0666))
}
func genClass(tableDesc *desc.MessageDescriptor) {
	name := tableDesc.GetName()
	if _, ok := classes[name]; !ok {
		cls := class{
			TypeName: name,
			Fields:   nil,
		}
		descFields := tableDesc.GetFields()
		for _, descField := range descFields {
			fieldType := protoToCppType(descField)
			fed := field{
				TypeName:  fieldType,
				FieldName: descField.GetName(),
			}
			cls.Fields = append(cls.Fields, fed)
		}
		classes[name] = cls
	}
}

func protoToCppType(descField *desc.FieldDescriptor) string {
	typ := descField.GetType()
	switch typ {
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		if descField.IsRepeated() {
			return "std::vector<int64_t>"
		}
		return "int64_t"
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		if descField.IsMap() {
			key := protoToCppType(descField.GetMapKeyType())
			value := protoToCppType(descField.GetMapValueType())
			return fmt.Sprintf("std::unordered_map<%s,%s>", key, value)
		} else if descField.IsRepeated() {
			rmt := descField.GetMessageType()
			genClass(rmt)
			return fmt.Sprintf("std::vector<%s>", rmt.GetName())
		} else {
			genClass(descField.GetMessageType())
			return descField.GetMessageType().GetName()
		}
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		if descField.IsRepeated() {
			return "std::vector<std::string>"
		}
		return "std::string"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		if descField.IsRepeated() {
			return "std::vector<bool>"
		}
		return "bool"
	default:
		panic(typ.String())
	}
	return ""
}
