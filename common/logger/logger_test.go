package logger

import (
	"log"
	"net"
	"testing"
	"time"

	"go.uber.org/zap"

	"go.uber.org/zap/zapcore"
)

func TestMustNew(t *testing.T) {
	//go logServer()
	time.Sleep(time.Second)
	l := MustNew(zapcore.DebugLevel, &Options{
		//Console: "stdout",
		//FilePath:     []string{"./testlog.log"},
		//RemoteAddr:   []string{"127.0.0.1:8081"},
		InitFields:   map[string]interface{}{"server_id": 1},
		Development:  true,
		RootCallPath: "easydev",
	})
	l.Debug("test1 debug", zap.String("test", "test"))
	l.Info("test1 info", zap.String("test", "test"))
	l.Warn("test1 warn", zap.String("test", "test"))
	l.Error("test1 error", zap.String("test", "test"))
	l1 := l.With(zap.Fields(zap.Bool("with", true)))
	l1.Debug("test2 debug", zap.String("test", "test"))
	l2 := l1.With(zap.AddCallerSkip(0))
	for i := 0; i < 100000; i++ {
		l2.Info("l2 info", zap.String("test3", "test"), zap.Time("time", time.Now()), zap.Int("index", i))
	}
	l.Sync()
	time.Sleep(time.Second * 3)
}

func logServer() {

	addr := "0.0.0.0:8081"

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)

	if err != nil {
		log.Fatalf("net.ResovleTCPAddr fail:%s", addr)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("listen %s fail: %s", addr, err)
	} else {

		log.Println("tcp listening", addr)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("listener.Accept error:", err)
			continue
		}

		go handleConnection(conn)

	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	data := make([]byte, 4096)
	for {
		n, err := conn.Read(data)
		if err != nil {
			panic(err)
		}
		log.Println("------------tcp:", string(data[:n]), "------------------")
	}
}
