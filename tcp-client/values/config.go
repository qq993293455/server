package values

import (
	"net"
	"os"
)

type ServerConfig struct {
	ServerName  string `json:"name"`
	GateWayAddr string `json:"gateAddr"`
	ServerId    int64  `json:"server_id"`
	StatelessId int64  `json:"stateless_id"`
	BattleId    int64  `json:"battle_id"`
}

type RequestJson struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SendRequest struct {
	Server string `json:"server"`
	User   string `json:"user"`
	Proto  string `json:"proto"`
}

type EnablePush struct {
	User string `json:"user"`
	Push bool   `json:"push"`
}

const (
	AppKey            = "app_key"
	Language          = 1
	RuleVersion       = ""
	Version           = 0
	LoginProtoMsgName = "service.User.RoleLoginRequest"
)

var (
	Users        = make([]string, 0)
	ServerCfg    = make([]ServerConfig, 0)
	JsonFilePath = "tcp-client/request.json"
	ProtoPath    = "."
	MAC          = ""
	Push         = true
)

func SetWorkingDir(dir string) {
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func GetMAC() {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Poor soul, here is what you got: " + err.Error())
	}
	//for _, inter := range interfaces {
	//fmt.Println(inter.Name)
	for _, address := range interfaces {
		if address.HardwareAddr != nil {
			MAC = address.HardwareAddr.String() //获取本机MAC地址
			break
		}
	}
}
