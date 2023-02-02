package main

import (
	"coin-server/common/logger"
	"fmt"
	"path/filepath"

	proto2 "coin-server/check_message/proto"
	"coin-server/check_message/values"
	"coin-server/common/proto/models"

	"go.uber.org/zap"
)

func main() {
	InitLog()
	initDefaultPath()
	//quit
	// quit := make(chan os.Signal)
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// <-quit
	// log.Println("Shutting down server...")
	// log.Println("Write json file finish")
}

func InitLog() {
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		FilePath:   []string{fmt.Sprintf("./%s.log", models.ServerType_LoadTest.String())},
		RemoteAddr: nil,
		InitFields: map[string]interface{}{
			"serverType": models.ServerType_LoadTest,
			"serverId":   values.ServerId,
		},
		Development: true,
		Discard:     true,
	})
	logger.SetDefaultLogger(log)
}

func initDefaultPath() {
	absPath, err := filepath.Abs(values.ProtoDir)
	if err != nil {
		panic(err)
	}
	values.SetWorkingDir(absPath)
	pbPkg := proto2.GetProtoPackage()
	pbPkg.InitProto(values.ProtoPath)
	pbPkg.InstanceProto(values.GateServerAddr, int64(values.ServerId))
}
