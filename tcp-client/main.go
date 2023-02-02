package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"coin-server/tcp-client/cli"
	"coin-server/tcp-client/nutsdb"
	_ "coin-server/tcp-client/nutsdb"
	proto2 "coin-server/tcp-client/proto"
	"coin-server/tcp-client/values"
)

func main() {
	initDefaultPath()
	cli.StartClientServer()
	openBrowser("http://localhost:8080")
	//quit
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	proto2.WriteJson(values.JsonFilePath)
	log.Println("Write json file finish")
}

func initDefaultPath() {
	d := nutsdb.Db.Get(values.MAC)
	if d != nil {
		return
	}
	dir := "../share/proto"
	absPath, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	values.SetWorkingDir(absPath)
	proto2.InitProto(values.ProtoPath)
	err = nutsdb.Db.Set(values.MAC, []byte(absPath))
	if err != nil {
		panic(err)
	}
}

func openBrowser(uri string) {
	var run string
	var args []string
	switch runtime.GOOS {
	case "windows":
		run = "cmd.exe"
		args = append(args, "/c")
		args = append(args, "start")
		args = append(args, uri)
	case "darwin":
		run = "open"
		args = append(args, uri)
	case "linux":
		run = "xdg-open"
		args = append(args, uri)
	default:
		fmt.Printf("don't know how to open things on %s platform", runtime.GOOS)
	}

	cmd := exec.Command(run, args...)
	_ = cmd.Start()
}
