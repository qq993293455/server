package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"coin-server/common/logger"
	"coin-server/common/msgcreate"
	"coin-server/common/network/stdtcp"
	_ "coin-server/common/proto/dungeon_match"
	_ "coin-server/common/proto/gen_rank"
	_ "coin-server/common/proto/guild_filter_service"
	lessservicepb "coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	_ "coin-server/common/proto/recommend"
	_ "coin-server/common/proto/roguelike_match"
	_ "coin-server/common/proto/service"
	"coin-server/common/protocol"
	proto2 "coin-server/tcp-client/proto"
	"coin-server/tcp-client/values"

	"github.com/gogo/protobuf/proto"

	"github.com/gorilla/websocket"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	"go.uber.org/zap/zapcore"
)

var TcpCli = sync.Map{}
var userMap = sync.Map{}

type TcpClient interface {
	Close()
	GetSession() *stdtcp.Session
	SendRequest(msgName string)
	Send(msg proto.Message)
}

func WriteToClient(id string, msg string) {
	wsi, ok := userMap.Load(id)
	if ok {
		ws := wsi.(*websocket.Conn)
		ws.WriteMessage(websocket.TextMessage, []byte(msg))
	}
}

type tcpClient struct {
	sess *stdtcp.Session

	id           string
	close        chan bool
	server       values.ServerConfig
	logger       *logger.Logger
	traceIdMutex sync.RWMutex
	response     map[uint32]chan *models.Resp
	traceId      string
	once         sync.Once
	timer        *time.Timer
}

func NewClient(uid string, server values.ServerConfig) (cli TcpClient, err error) {
	c := make(chan bool, 1)

	loger := logger.MustNew(zapcore.DebugLevel, &logger.Options{
		Console:     "stdout",
		FilePath:    nil,
		RemoteAddr:  nil,
		InitFields:  nil,
		Development: true,
	})

	t := &tcpClient{
		close:        c,
		id:           uid,
		server:       server,
		logger:       loger,
		traceIdMutex: sync.RWMutex{},
		traceId:      "traceId",
		response:     make(map[uint32]chan *models.Resp),
		timer:        time.NewTimer(time.Minute),
	}

	stdtcp.Connect(server.GateWayAddr, time.Second*3, false, t, loger, false)
	return t, nil
}

func InitClient(id string, server values.ServerConfig) {
	cli, err := NewClient(id, server)
	if err != nil {
		panic(err)
	}
	TcpCli.Store(id, cli)
}

func (t *tcpClient) Close() {
	t.sess.Close(nil)
}

func (t *tcpClient) GetSession() *stdtcp.Session {
	return t.sess
}

func (t *tcpClient) Send(msg proto.Message) {
	t.sess.Send(nil, msg)
}

func (t *tcpClient) SendRequest(msgName string) {
	// 请求json
	js := proto2.GetRequestJson(msgName)
	if js == nil {
		proto2.SaveRequestJson(msgName, proto2.PbToJson(msgName))
		js = proto2.GetRequestJson(msgName)
	}
	// 基于json构造pb
	req := msgcreate.NewMessage(msgName)
	unmarshaler := jsonpb.Unmarshaler{AllowUnknownFields: false, AnyResolver: nil}
	if err := unmarshaler.Unmarshal(bytes.NewReader(js), req); err != nil {
		err = errors.New(fmt.Sprintf("request.json 信息错误,expected:", msgName))
		panic(err)
	}

	// 返回pb
	out := &models.Resp{}
	err := t.sess.RPCRequest(nil, req, &out)
	if err != nil {
		WriteToClient(t.id, fmt.Sprintf("(请求失败) %s:", msgName))
		WriteToClient(t.id, fmt.Sprintf("ErrCode: %s", err.ErrCode))
		WriteToClient(t.id, fmt.Sprintf("ErrMsg: %s", err.ErrMsg))
		WriteToClient(t.id, fmt.Sprintf("ErrInfo: %s", err.ErrInternalInfo))
		return
	}
	// any转为response的pb
	if out.Resp == nil {
		WriteToClient(t.id, fmt.Sprintf("(请求出错) %s:", msgName))
		WriteToClient(t.id, fmt.Sprintf("ErrCode: %s", out.ErrCode))
		WriteToClient(t.id, fmt.Sprintf("ErrMsg: %s", out.ErrMsg))
		return
	}
	outMsg, err := protocol.AnyToMsg(out.Resp)
	if err != nil {
		err1 := errors.New(fmt.Sprintf("response message name not found:%s \n", out.Resp.GetTypeUrl()))
		panic(err1)
	}
	typeUrl := out.Resp.TypeUrl
	out.Resp.TypeUrl = "/" + out.Resp.TypeUrl
	err1 := types.UnmarshalAny(out.Resp, outMsg)
	if err1 != nil {
		panic(err1)
	}
	// pb转为json
	marshaller := jsonpb.Marshaler{}
	res, err1 := marshaller.MarshalToString(outMsg)
	if err1 != nil {
		panic(err1)
	}
	WriteToClient(t.id, typeUrl+":")
	WriteToClient(t.id, FormatJson(res))

	if out.OtherMsg != nil && values.Push {
		for _, msgAny := range out.OtherMsg {
			if msgName == (&models.PING{}).XXX_MessageName() {
				_ = t.sess.Send(nil, &models.PONG{})
				continue
			}
			msg, err2 := protocol.AnyToMsg(msgAny)
			if err2 != nil {
				err3 := errors.New(fmt.Sprintf("response message name not found:%s \n", out.Resp.GetTypeUrl()))
				panic(err3)
			}
			m := jsonpb.Marshaler{}
			p, err3 := m.MarshalToString(msg)
			if err3 != nil {
				panic(err3)
			}
			fmt.Println("receive push: ", p)
			WriteToClient(t.id, fmt.Sprintf("(推送) "+msgAny.GetTypeUrl()+":"))
			WriteToClient(t.id, FormatJson(p))
		}
	}
}

func (t *tcpClient) OnConnected(session *stdtcp.Session) {
	t.sess = session
	var out *models.Resp
	rl := &lessservicepb.User_RoleLoginRequest{
		UserId:        t.id,
		ServerId:      t.server.ServerId,
		AppKey:        values.AppKey,
		Language:      values.Language,
		RuleVersion:   values.RuleVersion,
		Version:       values.Version,
		BattleId:      t.server.BattleId,
		ClientVersion: "0.0.1",
	}
	err := session.RPCRequest(nil, rl, &out)
	if err != nil {
		WriteToClient(t.id, fmt.Sprintf("登陆失败"))
		WriteToClient(t.id, fmt.Sprintf("ErrCode: %s", err.ErrCode))
		WriteToClient(t.id, fmt.Sprintf("ErrMsg: %s", err.ErrMsg))
		WriteToClient(t.id, fmt.Sprintf("ErrInfo: %s", err.ErrInternalInfo))
		return
	}
	WriteToClient(t.id, fmt.Sprintf("登陆成功"))

	// 连战斗服
	//bp := &servicepb.GameBattle_EnterAreaRequest{
	//	Pos:            &battle.Pos{
	//		X: 199,
	//		Y: 199,
	//	},
	//	MapId:          "testMap_01",
	//	BattleServerId: t.server.BattleId,
	//}
	//out = &models.Resp{}
	//err = session.RPCRequest(nil, bp, &out)
	//if err != nil {
	//	WriteToClient(t.id, fmt.Sprintf("战斗服连接失败"))
	//	WriteToClient(t.id, fmt.Sprintf("ErrCode: %s", err.ErrCode))
	//	WriteToClient(t.id, fmt.Sprintf("ErrMsg: %s", err.ErrMsg))
	//	WriteToClient(t.id, fmt.Sprintf("ErrInfo: %s", err.ErrInternalInfo))
	//	return
	//}
	//WriteToClient(t.id, fmt.Sprintf("战斗服连接成功"))
}

func (t *tcpClient) OnDisconnected(session *stdtcp.Session, err error) {
	WriteToClient(t.id, fmt.Sprintf("与服务器断开连接"))
	userMap.Delete(t.id)
	TcpCli.Delete(t.id)
}

func (t *tcpClient) OnRequest(session *stdtcp.Session, rpcIndex uint32, msgName string, frame []byte) {
}

func (t *tcpClient) OnMessage(session *stdtcp.Session, msgName string, frame []byte) {
	if !values.Push {
		return
	}
	h := &models.ServerHeader{}
	out := &models.Resp{}
	err := protocol.DecodeInternal(frame, h, out)
	if err != nil {
		panic(err)
	}
	if msgName == (&models.PING{}).XXX_MessageName() {
		_ = session.Send(nil, &models.PONG{})
		return
	}

	for _, msgAny := range out.OtherMsg {
		msg, err1 := protocol.AnyToMsg(msgAny)
		if err1 != nil {
			err2 := errors.New(fmt.Sprintf("response message name not found:%s \n", out.Resp.GetTypeUrl()))
			panic(err2)
		}
		marshaller := jsonpb.Marshaler{}
		js, err2 := marshaller.MarshalToString(msg)
		if err2 != nil {
			panic(err2)
		}
		fmt.Println("receive push: ", js)
		WriteToClient(t.id, fmt.Sprintf("(推送) "+msgAny.GetTypeUrl()+":"))
		WriteToClient(t.id, FormatJson(js))
	}
}

func FormatJson(data string) string {
	var str bytes.Buffer
	_ = json.Indent(&str, []byte(data), "", "    ")
	return str.String()
}
