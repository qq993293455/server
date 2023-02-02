package main

import (
	"io/ioutil"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	plugin "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"github.com/gogo/protobuf/vanity"
	"github.com/gogo/protobuf/vanity/command"
)

func Read() *plugin.CodeGeneratorRequest {
	g := generator.New()
	data, err := ioutil.ReadFile("./testpb")
	if err != nil {
		g.Error(err, "reading input")
	}

	if err := proto.Unmarshal(data, g.Request); err != nil {
		g.Error(err, "parsing input proto")
	}

	if len(g.Request.FileToGenerate) == 0 {
		g.Fail("no files to generate")
	}
	return g.Request
}

func main() {
	req := command.Read()
	files := req.GetProtoFile()
	// time.Sleep(time.Second * 10)
	files = vanity.FilterFiles(files, vanity.NotGoogleProtobufDescriptorProto)
	generator.RegisterPlugin(&createMessage{})
	vanity.ForEachFile(files, vanity.TurnOnMarshalerAll)
	vanity.ForEachFile(files, vanity.TurnOnSizerAll)
	vanity.ForEachFile(files, vanity.TurnOnUnmarshalerAll)

	vanity.ForEachFieldInFilesExcludingExtensions(vanity.OnlyProto2(files), vanity.TurnOffNullableForNativeTypesWithoutDefaultsOnly)
	vanity.ForEachFile(files, vanity.TurnOffGoUnrecognizedAll)
	vanity.ForEachFile(files, vanity.TurnOffGoUnkeyedAll)
	vanity.ForEachFile(files, vanity.TurnOffGoSizecacheAll)

	//vanity.ForEachFile(files, vanity.TurnOffGoEnumPrefixAll)
	vanity.ForEachFile(files, vanity.TurnOffGoEnumStringerAll)
	vanity.ForEachFile(files, vanity.TurnOnEnumStringerAll)

	vanity.ForEachFile(files, vanity.TurnOnEqualAll)
	//vanity.ForEachFile(files, vanity.TurnOffGoStringAll)
	vanity.ForEachFile(files, vanity.TurnOffGoStringerAll)
	// vanity.ForEachFile(files, vanity.TurnOnStringerAll)
	vanity.ForEachFile(files, vanity.TurnOnMessageNameAll)
	vanity.ForEachFile(files, TurnOnNewMessageAll)

	resp := command.Generate(req)
	command.Write(resp)
}

func TurnOnNewMessageAll(file *descriptor.FileDescriptorProto) {
	vanity.SetBoolFileOption(gogoproto.E_MessagenameAll, true)(file)
}

var E_NewMessageAll = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.FileOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         64033,
	Name:          "gogoproto.newmessage_all",
	Tag:           "varint,64033,opt,name=newmessage_all",
	Filename:      "gogo.proto",
}
