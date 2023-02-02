package cpb

import (
	"coin-server/common/utils"

	"github.com/jhump/protoreflect/desc/builder"
	"github.com/jhump/protoreflect/desc/protoprint"
)

func Write(b *builder.FileBuilder, rootDir string) {
	//p := &protoprint.Printer{}
	//fileDesc, err := b.Build()
	//utils.Must(err)
	//err = p.PrintProtosToFileSystem([]*desc.FileDescriptor{fileDesc}, rootDir)
	//utils.Must(err)
}

func ToString(b *builder.FileBuilder) string {
	p := &protoprint.Printer{}
	fileDesc, err := b.Build()
	utils.Must(err)
	str, err := p.PrintProtoToString(fileDesc)
	utils.Must(err)
	return str
}
