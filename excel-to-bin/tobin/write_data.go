package tobin

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"coin-server/common/utils"
	"coin-server/excel-to-bin/bufferstream"
	"coin-server/excel-to-bin/parse"
	"coin-server/rule"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/tidwall/gjson"
)

func assert(b bool) {
	if !b {
		panic("assert failed")
	}
}

func WriteData(fb *builder.FileBuilder, dataS []rule.TableData, tableInfos map[string]*parse.TableInfo, pathS []string) {
	dataMap := map[string]string{}
	for _, v := range dataS {
		dataMap[v.Table] = v.Data
	}
	fd, err := fb.Build()
	utils.Must(err)
	tablesDesc := fd.FindMessage("configcpp.Tables")
	dMsg := dynamic.NewMessage(tablesDesc)
	fs := tablesDesc.GetFields()
	for _, v := range fs {
		if !v.IsMap() {
			if v.GetName() == "bin_data_version" {
				continue
			} else {
				panic("Tables field must is map")
			}
		}

		data, ok := dataMap[v.GetName()]
		if !ok {
			panic("not found data:" + v.GetName())
		}
		msgType := v.GetMapValueType().GetMessageType()
		retS := gjson.Parse(data).Array()
		for _, ret := range retS {
			fMsg := parseMessage(msgType, ret)
			if msgType.GetName() == "KeyValue" {
				idString := ret.Get("key").String()
				assert(idString != "")
				dMsg.PutMapField(v, idString, fMsg)
			} else {
				idInt64 := ret.Get("id").Int()
				assert(idInt64 != 0)
				dMsg.PutMapField(v, idInt64, fMsg)
			}
		}
	}
	versionDesc := tablesDesc.FindFieldByName("bin_data_version")
	dMsg.SetField(versionDesc, time.Now().UnixNano())
	for _, v := range tableInfos {
		if v.Child != "" {
			dropMiniName := v.Child
			dropMini := dataMap[dropMiniName]
			dmnDesc := fd.FindMessage("configcpp." + utils.ToCamel(dropMiniName))
			parentV := dMsg.GetFieldByName(v.Name)
			parentI := parentV.(map[interface{}]interface{})
			retS := gjson.Parse(dropMini).Array()
			for _, ret := range retS {
				msg := parseMessage(dmnDesc, ret)
				key := msg.GetFieldByNumber(1).(int64)
				childKey := msg.GetFieldByNumber(2)
				v := parentI[key]
				assert(v != nil)
				v.(*dynamic.Message).PutMapFieldByName(dropMiniName, childKey, msg)
			}
		}
	}
	bs := &bufferstream.BaseStream{}
	bs.WriteType(bufferstream.CustomStructT)
	fields := tablesDesc.GetFields()
	for _, field := range fields {
		f := dMsg.GetField(field)
		bufferstream.Write(bs, f)
	}
	bufferstream.GenCppClass(tablesDesc)
	utils.Must(err)
	for _, v := range pathS {
		utils.Must(ioutil.WriteFile(v, bs.Bytes(), 666))
	}
}

func parseMessage(msgType *desc.MessageDescriptor, ret gjson.Result) *dynamic.Message {
	fMsg := dynamic.NewMessage(msgType)

	mts := msgType.GetFields()
	for _, f := range mts {
		switch f.GetType() {
		case descriptor.FieldDescriptorProto_TYPE_INT64:
			if f.IsRepeated() {
				var slice []int64
				for _, v := range ret.Get(f.GetName()).Array() {
					slice = append(slice, v.Int())
				}
				fMsg.SetFieldByName(f.GetName(), slice)
			} else {
				int64t := ret.Get(f.GetName()).Int()
				fMsg.SetFieldByName(f.GetName(), int64t)
			}
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			if f.IsRepeated() {
				var slice []string
				for _, v := range ret.Get(f.GetName()).Array() {
					slice = append(slice, v.String())
				}
				fMsg.SetFieldByName(f.GetName(), slice)
			} else {
				str := ret.Get(f.GetName()).String()
				fMsg.SetFieldByName(f.GetName(), str)
			}
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			fMsg.SetFieldByName(f.GetName(), ret.Get(f.GetName()).Bool())
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			//name := f.GetName()
			if f.IsMap() {
				if f.GetMapValueType().GetType() != descriptor.FieldDescriptorProto_TYPE_MESSAGE {
					mapRet := ret.Get(f.GetName()).Map()
					switch f.GetMapKeyType().GetType() {
					case descriptor.FieldDescriptorProto_TYPE_INT64:
						switch f.GetMapValueType().GetType() {
						case descriptor.FieldDescriptorProto_TYPE_INT64:
							for k, v := range mapRet {
								key, err := strconv.Atoi(k)
								utils.Must(err)
								fMsg.PutMapFieldByName(f.GetName(), key, v.Int())
							}

						case descriptor.FieldDescriptorProto_TYPE_STRING:
							for k, v := range mapRet {
								key, err := strconv.Atoi(k)
								utils.Must(err)
								fMsg.PutMapFieldByName(f.GetName(), key, v.String())
							}
						default:
							panic(fmt.Sprintf("unknown field:%v:%v", f.GetType(), f.GetName()))
						}
					case descriptor.FieldDescriptorProto_TYPE_STRING:
						switch f.GetMapValueType().GetType() {
						case descriptor.FieldDescriptorProto_TYPE_INT64:
							for k, v := range mapRet {
								fMsg.PutMapFieldByName(f.GetName(), k, v.Int())
							}
						case descriptor.FieldDescriptorProto_TYPE_STRING:
							for k, v := range mapRet {
								fMsg.PutMapFieldByName(f.GetName(), k, v.String())
							}
						default:
							panic(fmt.Sprintf("unknown field:%v:%v", f.GetType(), f.GetName()))
						}
					default:
						panic(fmt.Sprintf("unknown field:%v:%v", f.GetType(), f.GetName()))
					}
				}
			} else {
				if f.GetMessageType().GetName() == "DoubleRepeatedInt64" && f.IsRepeated() {
					for _, ar := range ret.Get(f.GetName()).Array() {
						var arr []int64
						for _, ar1 := range ar.Array() {
							arr = append(arr, ar1.Int())
						}
						slice := dynamic.NewMessage(f.GetMessageType())
						slice.SetFieldByNumber(1, arr)
						fMsg.AddRepeatedFieldByName(f.GetName(), slice)
					}
				} else {
					panic(fmt.Sprintf("unknown field:%v:%v", f.GetType(), f.GetName()))
				}
			}

		default:
			panic(fmt.Sprintf("unknown field:%v:%v", f.GetType(), f.GetName()))
		}
	}
	return fMsg
}
