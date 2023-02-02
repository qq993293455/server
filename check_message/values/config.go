package values

import "os"

var (
	GateServerAddr = "10.11.195.229:8071"
	ProtoDir       = "../../share/proto/"
	ProtoPath      = "./proto"
	ServerId       = 1
)

func SetWorkingDir(dir string) {
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}
