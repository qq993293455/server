package nutsdb

import (
	"encoding/json"

	proto2 "coin-server/tcp-client/proto"
	"coin-server/tcp-client/values"
)

var Db *NutsDb

func init() {
	err := NewNutsDb().Connect()
	if err != nil {
		panic(err)
	}

	values.GetMAC()
	dir := Db.Get(values.MAC)
	if dir == nil {
		return
	}
	values.SetWorkingDir(string(dir))
	proto2.InitProto(values.ProtoPath)
	serverCfg := Db.Get("server")
	if serverCfg == nil {
		return
	}
	err = json.Unmarshal(serverCfg, &values.ServerCfg)
	if err != nil {
		panic(err)
	}
}
