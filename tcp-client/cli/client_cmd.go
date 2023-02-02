package cli

import (
	"fmt"
	"sync/atomic"
	"time"

	"coin-server/common/logger"
	"coin-server/common/msgcreate"
	"coin-server/common/network/stdtcp"
	"coin-server/common/pprof"
	lessservicepb "coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	"coin-server/common/protocol"

	"go.uber.org/zap/zapcore"
)

var count uint64

type Client struct {
	log  *logger.Logger
	sess *stdtcp.Session
}

func NewCmdClient(addr string, log *logger.Logger) *Client {
	c := &Client{
		log: log,
	}

	stdtcp.Connect(addr, time.Second*3, true, c, log, false)
	return c
}

//var userId = time.Now().UnixNano()
var userId = int64(1000000)

func Run() {
	log := logger.MustNew(zapcore.DebugLevel, &logger.Options{
		Console:     "stdout",
		FilePath:    nil,
		RemoteAddr:  nil,
		InitFields:  nil,
		Development: true,
	})
	go func() {
		for {
			time.Sleep(time.Second)
			fmt.Println(atomic.LoadUint64(&count))
		}
	}()
	pprof.Start(log, ":6011")
	m := map[*Client]struct{}{}
	for i := 0; i < 1; i++ {
		c := NewCmdClient("127.0.0.1:8071", log)
		m[c] = struct{}{}
		time.Sleep(time.Millisecond)
	}
	for {
		time.Sleep(time.Second)
	}
}

func (this_ *Client) OnConnected(session *stdtcp.Session) {
	this_.sess = session
	var out *models.Resp
	rl := &lessservicepb.User_RoleLoginRequest{
		UserId:        "czy1",
		ServerId:      11,
		AppKey:        "",
		Language:      1,
		RuleVersion:   "",
		Version:       0,
		ClientVersion: "0.0.1",
	}
	err := session.RPCRequest(nil, rl, &out)
	if err != nil {
		panic(err)
	}
}

func (this_ *Client) OnDisconnected(session *stdtcp.Session, err error) {

}

func (this_ *Client) OnRequest(session *stdtcp.Session, rpcIndex uint32, msgName string, frame []byte) {

}

func (this_ *Client) OnMessage(session *stdtcp.Session, msgName string, frame []byte) {
	h := &models.ServerHeader{}
	msg := msgcreate.NewMessage(msgName)
	err := protocol.DecodeInternal(frame, h, msg)
	if err != nil {
		panic(err)
	}
	// fmt.Println(h, msg)
	if msgName == (&models.PING{}).XXX_MessageName() {
		_ = session.Send(nil, &models.PONG{})
	}

}

func (this_ *Client) Close() {
	this_.sess.Close(nil)
}
