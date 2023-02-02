package proto

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

type MessagePair struct {
	ProtoName    string
	Request      *desc.MessageDescriptor
	Response     *desc.MessageDescriptor
	Comment      string
	RequestName  string
	ResponseName string
}

type SourceProto struct {
	MessageMap map[string]*MessagePair
}

func NewSourceProto() *SourceProto {
	return &SourceProto{
		MessageMap: map[string]*MessagePair{}}
}

func getProtoFiles(srcPath string) []string {
	var outPaths []string
	err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		Must(err)
		if info.IsDir() {
			return nil
		}
		if path[:10] == "clientutil" {
			return nil
		}
		if filepath.Ext(path) != ".proto" {
			return nil
		}
		path = strings.ReplaceAll(path, "\\", "/")
		outPaths = append(outPaths, path)
		return nil
	})
	Must(err)
	return outPaths
}

func getProtoDir(srcPath string) []string {
	var outPaths []string
	err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		Must(err)
		if !info.IsDir() {
			return nil
		}
		absPath, err := filepath.Abs(path)
		Must(err)
		//防止重名
		if !strings.Contains(absPath, "dao") {
			outPaths = append(outPaths, absPath)
		}
		return nil
	})
	Must(err)

	return outPaths
}

func (this_ *SourceProto) ParseProtoFiles(path string) {
	files := getProtoFiles(path)
	//fmt.Println(protoparse.ResolveFilenames([]string{}, files...))
	parse := &protoparse.Parser{
		ValidateUnlinkedFiles: true,
		InferImportPaths:      true,
		IncludeSourceCodeInfo: true,
	}
	fds, err := parse.ParseFiles(files...)
	Must(err)
	messageMap := map[string]*MessagePair{}
	for _, fd := range fds {
		msgS := fd.GetMessageTypes()
		for _, msg := range msgS {
			protoName := msg.GetFullyQualifiedName()
			msgComment := msg.GetSourceInfo().GetLeadingComments()
			if len(msg.GetNestedMessageTypes()) == 0 {
				mp, ok := messageMap[protoName]
				if !ok {
					mp = &MessagePair{RequestName: "", ResponseName: ""}
					messageMap[protoName] = mp
				}
				if mp.RequestName != "" {
					panic(fmt.Sprintf("dumplicate error:%s had response name %s,please change id of %s", protoName, mp.Response.GetName(), protoName))
				}
				mp.ProtoName = protoName
				mp.RequestName = protoName
				mp.Request = msg
				mp.Comment = msgComment
			}
			for _, nestMsg := range msg.GetNestedMessageTypes() {
				msgName := nestMsg.GetName()
				name := fmt.Sprintf("%s.%s", protoName, msgName)
				mp, ok := messageMap[name]
				if !ok {
					mp = &MessagePair{RequestName: "", ResponseName: ""}
					messageMap[name] = mp
				}
				if strings.Contains(name, "Request") {
					if mp.RequestName != "" {
						panic(fmt.Sprintf("dumplicate error:%s had request name %s,please change id of %s", protoName, mp.Request.GetName(), name))
					}
					mp.RequestName = msgName
					mp.ProtoName = name
					mp.Request = nestMsg
					mp.Comment = msgComment
				} else {
					if mp.RequestName != "" {
						panic(fmt.Sprintf("dumplicate error:%s had response name %s,please change id of %s", protoName, mp.Response.GetName(), name))
					}
					mp.ResponseName = msgName
					mp.Response = nestMsg
					mp.ProtoName = name
					if mp.Comment == "" {
						mp.Comment = msgComment
					}
				}
			}
		}
	}
	this_.MessageMap = messageMap
}

func (this_ *SourceProto) GetProtoMessageByName(name string) (*MessagePair, bool) {
	v, ok := this_.MessageMap[name]
	return v, ok
}

func (this_ *SourceProto) GetRequestNames() []string {
	str := make([]string, 0)
	for k := range this_.MessageMap {
		if strings.Contains(k, "Request") {
			str = append(str, k)
		}
	}
	return str
}

func (this_ *SourceProto) NewRequestMsgByName(name string) (*dynamic.Message, bool) {
	mp, ok := this_.GetProtoMessageByName(name)
	if !ok {
		return nil, false
	}
	return dynamic.NewMessage(mp.Request), true
}

func (this_ *SourceProto) NewResponseMsgByName(name string) (*dynamic.Message, bool) {
	mp, ok := this_.GetProtoMessageByName(name)
	if !ok {
		return nil, false
	}
	return dynamic.NewMessage(mp.Response), true
}

func (this_ *SourceProto) NewPushMsgByName(name string) (*dynamic.Message, bool) {
	mp, ok := this_.GetProtoMessageByName(name)
	if !ok {
		return nil, false
	}
	return dynamic.NewMessage(mp.Response), true
}

func (this_ *SourceProto) GetComment(name string) string {
	mp, ok := this_.GetProtoMessageByName(name)
	if !ok {
		return ""
	}
	return strings.Split(mp.Comment, "\n")[0]
}
