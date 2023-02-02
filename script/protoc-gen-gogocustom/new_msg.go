package main

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	descriptorpb "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

type createMessage struct {
	*generator.Generator
	generator.PluginImports
}

func (p *createMessage) Name() string {
	return "createMessage"
}

func (p *createMessage) Init(g *generator.Generator) {
	p.Generator = g
}

func (p *createMessage) Generate(file *generator.FileDescriptor) {
	p.PluginImports = generator.NewPluginImports(p.Generator)
	msgtype := p.NewImport("coin-server/common/msgcreate")
	strconvPKG := p.NewImport("strconv")
	bytespool := p.NewImport("coin-server/common/bytespool")
	sync := p.NewImport("sync")
	jsonPkG := p.NewImport("github.com/json-iterator/go")
	p.P("func init(){")
	for _, message := range file.Messages() {
		if message.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}
		p.In()
		p.P(msgtype.Use(), ".RegisterNewMessage(", fmt.Sprintf(funcStr, strings.Join(message.TypeName(), "_")), ")")
		p.Out()
	}
	p.P("}")
	if len(file.Messages()) > 0 {
		sync.Use()
	}
	for _, message := range file.Messages() {
		if message.DescriptorProto.GetOptions().GetMapEntry() {
			continue
		}
		msgName := strings.Join(message.TypeName(), "_")
		p.P(fmt.Sprintf(poolStr, msgName, msgName))
		p.P(fmt.Sprintf(poolPutStr, msgName, msgName))
	}
	for _, message := range file.Messages() {
		for _, v := range message.Field {
			if v.Options != nil {
				o, err := proto.GetExtension(v.Options, gogoproto.E_Moretags)
				if err == nil && o.(*string) != nil {
					str := *(o.(*string))
					if str == "pk" {
						t := v.GetType()
						switch t {
						case descriptorpb.FieldDescriptorProto_TYPE_STRING:
							p.P(fmt.Sprintf(pkStringStr, message.GetName(), generator.CamelCase(v.GetName()), message.GetName(), generator.CamelCase(v.GetName())))
						case
							descriptorpb.FieldDescriptorProto_TYPE_INT64,
							descriptorpb.FieldDescriptorProto_TYPE_INT32,
							descriptorpb.FieldDescriptorProto_TYPE_SINT32,
							descriptorpb.FieldDescriptorProto_TYPE_SINT64,
							descriptorpb.FieldDescriptorProto_TYPE_ENUM,
							descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
							descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
							strconvPKG.Use()
							p.P(fmt.Sprintf(pkIntStr, message.GetName(), generator.CamelCase(v.GetName()), message.GetName(), generator.CamelCase(v.GetName())))
						case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
							strconvPKG.Use()
							p.P(fmt.Sprintf(pkBoolStr, message.GetName(), generator.CamelCase(v.GetName()), message.GetName(), generator.CamelCase(v.GetName())))
						case
							descriptorpb.FieldDescriptorProto_TYPE_UINT32,
							descriptorpb.FieldDescriptorProto_TYPE_UINT64,
							descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
							descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
							strconvPKG.Use()
							p.P(fmt.Sprintf(pkUintStr, message.GetName(), generator.CamelCase(v.GetName()), message.GetName(), generator.CamelCase(v.GetName())))
						case
							descriptorpb.FieldDescriptorProto_TYPE_DOUBLE,
							descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
							strconvPKG.Use()
							p.P(fmt.Sprintf(pkFloatStr, message.GetName(), generator.CamelCase(v.GetName()), message.GetName(), generator.CamelCase(v.GetName())))
						default:
							panic("pk not support :" + t.String())
						}
						p.P(fmt.Sprintf(toSaveStr, message.GetName(), bytespool.Use(), message.GetName(), bytespool.Use(), message.GetName()))
					}
					if str == "rpt_int64" {
						jsonPkG.Use()
						p.P(fmt.Sprintf(rptInt64Str, message.GetName()))
					}
				}
			}
		}
	}
}

var pkStringStr = `
func (m*%s)PK() string{
	if m==nil{
		return ""	
	}
	return m.%s
}

func (m*%s)PKAppendTo(d []byte)[]byte{
	if m==nil{
		return d	
	}
	return append(d,m.%s...)
}
`

var pkIntStr = `
func (m*%s)PK() string{
	if m==nil{
		return ""	
	}
	return strconv.FormatInt(int64(m.%s), 10)
}

func (m *%s) PKAppendTo(d []byte) []byte {
	if m == nil {
		return d
	}
	return strconv.AppendInt(d, int64(m.%s), 10)
}

`

var pkBoolStr = `
func (m*%s)PK() string{
	if m==nil{
		return ""	
	}
	return strconv.FormatBool(bool(m.%s))
}

func (m *%s) PKAppendTo(d []byte) []byte {
	if m == nil {
		return d
	}
	return strconv.AppendBool(d, bool(m.%s))
}
`

var pkUintStr = `
func (m*%s)PK() string{
	if m==nil{
		return ""	
	}
	return strconv.FormatUint(uint64(m.%s),10)
}

func (m *%s) PKAppendTo(d []byte) []byte {
	if m == nil {
		return d
	}
	return strconv.AppendUint(d, uint64(m.%s), 10)
}

`

var pkFloatStr = `
func (m*%s)PK() string{
	if m==nil{
		return ""	
	}
	return strconv.FormatFloat(float64(m.%s), 'g', -1, 64)
}

func (m *%s) PKAppendTo(d []byte) []byte {
	if m == nil {
		return d
	}
	return strconv.AppendFloat(d, float64(m.%s), 'g', -1, 64)
}

`
var toSaveStr = `
func (m *%s) ToKVSave() ([]byte, []byte) {
	msgName := m.XXX_MessageName()
	dk := %s.GetSample(64)
	dk = dk[:0]
	dk = append(dk, msgName...)
	dk = append(dk, ':', 'k', ':')
	dk = m.PKAppendTo(dk)
	return dk, m.ToSave()
}

func (m *%s) ToSave() []byte {
	msgName := m.XXX_MessageName()
	ml := len(msgName)
	d := %s.GetSample(1 + ml + m.Size())
	d[0] = uint8(ml)
	copy(d[1:], msgName)
	_, _ = m.MarshalToSizedBuffer(d[1+ml:])
	return d
}

func (m *%s) KVKey() string {
	return m.XXX_MessageName() + ":k:" + m.PK()
}
`

var funcStr = ` func() proto.Message{
		return pool%s.Get().(proto.Message)
	}`
var poolStr = `var pool%s = &sync.Pool{ New: func() interface{}{ return &%s{}}}`

var poolPutStr = `func (m*%s)ReleasePool()  { m.Reset(); pool%s.Put(m) ;m = nil }`

var rptInt64Str = `func (m *%s) UnmarshalJSON(data []byte) error {
	sls := make([]int64, 0)
	err := github_com_json_iterator_go.Unmarshal(data, &sls)
	if err != nil {
		return err
	}
	m.Slice = sls
	return nil
}
`
