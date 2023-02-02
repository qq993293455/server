package proto

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"coin-server/tcp-client/values"

	"github.com/fsnotify/fsnotify"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"
)

var ProtoRequestMap = make(map[string]proto.Message)
var ProtoResponseMap = make(map[string]proto.Message)

func init() {
	InitRequestJson(values.JsonFilePath)
}

func RegisterProto(protoMsg string, req proto.Message, resp proto.Message) {
	ProtoRequestMap[protoMsg] = req
	ProtoResponseMap[protoMsg] = resp
}
func InitConfig(path string) {
	viper.SetConfigType("json")
	viper.SetConfigName("request")
	viper.AddConfigPath(path)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("监听到文件发生了变化:", e)
	})
}

var protoSource atomic.Value
var requestJson atomic.Value

func genProtoSource(path string) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("panic Error:", e)
		} else {
			fmt.Println(time.Now(), "proto 信息已更新到最新")
		}
	}()
	p := NewSourceProto()
	p.ParseProtoFiles(path)
	protoSource.Store(p)
}

func Proto() *SourceProto {
	return protoSource.Load().(*SourceProto)
}

func InitProto(path string) {
	p := NewSourceProto()
	p.ParseProtoFiles(path)
	protoSource.Store(p)
	w, err := fsnotify.NewWatcher()
	Must(err)
	Must(w.Add(path))
	go func() {
		for {
			select {
			case ev := <-w.Events:
				{
					//判断事件发生的类型，如下5种
					// Create 创建
					// Write 写入
					// Remove 删除
					// Rename 重命名
					// Chmod 修改权限
					//if ev.Op&fsnotify.Create == fsnotify.Create {
					//	log.Println("创建文件 : ", ev.Name)
					//}
					//if ev.Op&fsnotify.Write == fsnotify.Write {
					//	log.Println("写入文件 : ", ev.Name)
					//}
					//if ev.Op&fsnotify.Remove == fsnotify.Remove {
					//	log.Println("删除文件 : ", ev.Name)
					//}
					//if ev.Op&fsnotify.Rename == fsnotify.Rename {
					//	log.Println("重命名文件 : ", ev.Name)
					//}
					//if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
					//	log.Println("修改权限 : ", ev.Name)
					//}
					if !strings.Contains(ev.Name, ".proto~") {
						time.Sleep(time.Second)
						genProtoSource(path)
					}

				}
			case err := <-w.Errors:
				{
					fmt.Println("error : ", err)
					return
				}
			}
		}
	}()
}

func GetRequestJson(key string) json.RawMessage {
	m := requestJson.Load().(map[string]json.RawMessage)
	return m[key]
}

func parseRequestJson(path string) (success bool) {
	path, err := filepath.Abs(path)
	Must(err)
	values.JsonFilePath = path
	m := map[string]json.RawMessage{}
	defer func() {
		if e := recover(); e != nil {
			log.Println("parse "+path+":", e)
			success = false
		} else {
			requestJson.Store(m)
			WriteJson(values.JsonFilePath)
			success = true
		}
	}()
	data, err := ioutil.ReadFile(path)
	Must(err)
	Must(json.Unmarshal(data, &m))
	return
}

func SaveRequestJson(key string, value string) (success bool) {
	_, ok := Proto().NewRequestMsgByName(key)
	if !ok {
		log.Printf("request message name not found:%s \n", key)
		return false
	}
	m := requestJson.Load().(map[string]json.RawMessage)
	defer func() {
		if e := recover(); e != nil {
			log.Println("parse: ", e)
			success = false
		} else {
			requestJson.Store(m)
			success = true
		}
	}()
	js := json.RawMessage{}
	Must(json.Unmarshal([]byte(value), &js))
	m[key] = js
	return
}

func InitRequestJson(path string) {
	if !parseRequestJson(path) {
		os.Exit(0)
	}
	w, err := fsnotify.NewWatcher()
	Must(err)
	Must(w.Add(path))
	go func() {
		for {
			select {
			case ev := <-w.Events:
				{
					//判断事件发生的类型，如下5种
					// Create 创建
					// Write 写入
					// Remove 删除
					// Rename 重命名
					// Chmod 修改权限
					//if ev.Op&fsnotify.Create == fsnotify.Create {
					//	log.Println("创建文件 : ", ev.Name)
					//}
					//if ev.Op&fsnotify.Write == fsnotify.Write {
					//	log.Println("写入文件 : ", ev.Name)
					//}
					//if ev.Op&fsnotify.Remove == fsnotify.Remove {
					//	log.Println("删除文件 : ", ev.Name)
					//}
					//if ev.Op&fsnotify.Rename == fsnotify.Rename {
					//	log.Println("重命名文件 : ", ev.Name)
					//}
					//if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
					//	log.Println("修改权限 : ", ev.Name)
					//}
					if ev.Name == path {
						parseRequestJson(path)
					}

				}
			case err := <-w.Errors:
				{
					fmt.Println("error : ", err)
					return
				}
			}
		}
	}()
}

func PbToJson(name string) string {
	req, ok := Proto().NewRequestMsgByName(name)
	if !ok {
		log.Printf("request message not found: %s \n", name)
		return ""
	}
	js, err := req.MarshalJSONPB(&jsonpb.Marshaler{OrigName: true, EmitDefaults: true})
	if err != nil {
		panic(err)
	}
	return string(js)
}

func LoginPbToJson(uid string, sid int64) string {
	req, ok := Proto().NewRequestMsgByName(values.LoginProtoMsgName)
	if !ok {
		log.Printf("request message not found: %s \n", values.LoginProtoMsgName)
		return ""
	}

	req.SetFieldByName("user_id", uid)
	req.SetFieldByName("server_id", sid)
	req.SetFieldByName("app_key", values.AppKey)
	req.SetFieldByName("language", values.Language)
	req.SetFieldByName("rule_version", values.RuleVersion)
	req.SetFieldByName("version", values.Version)
	js, err := req.MarshalJSONPB(&jsonpb.Marshaler{OrigName: true, EmitDefaults: true})
	if err != nil {
		panic(err)
	}

	return string(js)
}

func WriteJson(path string) {
	m := requestJson.Load().(map[string]json.RawMessage)

	js, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		fmt.Println("error : ", err)
	}
	err = ioutil.WriteFile(path, js, 0666)
	if err != nil {
		fmt.Println("error : ", err)
	}
}
