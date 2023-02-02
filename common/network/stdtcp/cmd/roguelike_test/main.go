package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"coin-server/common/logger"
	"coin-server/common/msgcreate"
	"coin-server/common/network/stdtcp"
	"coin-server/common/pprof"
	"coin-server/common/proto/cppbattle"
	lessservicepb "coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	rlpb "coin-server/common/proto/roguelike_match"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/protocol"
	"coin-server/common/values"
	"github.com/gogo/protobuf/proto"

	"go.uber.org/zap/zapcore"
)

var count uint64

type Client struct {
	log          *logger.Logger
	sess         *stdtcp.Session
	battleClient *Client
	type_        int64
	roleId       string
}

func NewClient(addr string, log *logger.Logger) *Client {
	c := &Client{
		log: log,
	}
	rand.Seed(time.Now().UnixNano())
	stdtcp.Connect(addr, time.Second*3, true, c, log, false)
	return c
}

var userIdPrefix = int64(287001002)
var userId = time.Now().UnixMilli() + userIdPrefix
var bossMap = map[values.Integer]values.Integer{
	1: 1,
	2: 3,
	3: 5,
	4: 7,
	5: 9,
	6: 11,
}

func (this_ *Client) OnConnected(session *stdtcp.Session) {
	if this_.type_ == 0 {
		this_.sess = session
		u := strconv.Itoa(int(atomic.AddInt64(&userId, 1)))
		//u := "ywhxx-05"
		rl := &lessservicepb.User_RoleLoginRequest{
			UserId:        u,
			ServerId:      1,
			AppKey:        "",
			Language:      0,
			RuleVersion:   "",
			Version:       0,
			ClientVersion: "0.0.1",
		}
		session.SetMeta(rl.UserId)
		respLogin := &lessservicepb.User_RoleLoginResponse{}
		err := session.RPCRequestOut(nil, rl, respLogin)
		if err != nil {
			fmt.Println("User_RoleLoginRequest painc")
			panic(err)
		}

		//respCenter := &servicepb.GameBattle_GetCurrBattleInfoResponse{}
		//err1 := session.RPCRequestOut(nil, &servicepb.GameBattle_GetCurrBattleInfoRequest{}, respCenter)
		//if err1 != nil {
		//	panic(err1)
		//}
		//fmt.Println("CurBattleInfoResponse:", respCenter)
		respAddLevelResp := &lessservicepb.User_CheatSetLevelResponse{}
		err = session.RPCRequestOut(nil, &lessservicepb.User_CheatSetLevelRequest{Level: 20}, respAddLevelResp)
		if err != nil {
			fmt.Println("User_CheatSetLevelRequest painc")
			panic(err)
		}

		respUnlockResp := &servicepb.SystemUnlock_CheatUnlockSystemResponse{}
		err = session.RPCRequestOut(nil, &servicepb.SystemUnlock_CheatUnlockSystemRequest{SystemId: models.SystemType_SystemPadding}, respUnlockResp)
		if err != nil {
			fmt.Println("SystemUnlock_CheatUnlockSystemRequest painc")
			panic(err)
		}

		respGetToday := &rlpb.RoguelikeMatch_RLGetTodayBossResponse{}
		err = session.RPCRequestOut(nil, &rlpb.RoguelikeMatch_RLGetTodayBossRequest{}, respGetToday)
		if err != nil {
			fmt.Println("RoguelikeMatch_RLGetTodayBossRequest painc")
			panic(err)
		}

		boss := bossMap[respGetToday.DungeonDay]
		if boss == 0 {
			panic(errors.New("no boss"))
		}

		respCreateRoom := &rlpb.RoguelikeMatch_RLCreateRoomResponse{}
		err = session.RPCRequestOut(nil, &rlpb.RoguelikeMatch_RLCreateRoomRequest{RoguelikeId: boss}, respCreateRoom)
		if err != nil {
			fmt.Println("RoguelikeMatch_RLCreateRoomRequest painc")
			panic(err)
		}

		respChooseHero := &rlpb.RoguelikeMatch_RLChooseHeroResponse{}
		err = session.RPCRequestOut(nil, &rlpb.RoguelikeMatch_RLChooseHeroRequest{ConfigId: 2001, CardId: 1}, respChooseHero)
		if err != nil {
			fmt.Println("RoguelikeMatch_RLChooseHeroRequest painc")
			panic(err)
		}

		respStartRL := &rlpb.RoguelikeMatch_RLStartBattleResponse{}
		err = session.RPCRequestOut(nil, &rlpb.RoguelikeMatch_RLStartBattleRequest{}, respStartRL)
		if err != nil {
			fmt.Println("RoguelikeMatch_RLStartBattleRequest painc")
			panic(err)
		}

	} else {
		push := cppbattle.NSNB_AuthTCPPushToServer{
			RoleId: this_.roleId,
			Token:  "",
		}
		err := session.Send(&models.ServerHeader{RoleId: this_.roleId}, &push)
		if err != nil {
			panic(err)
		}
	}
}

func (this_ *Client) OnDisconnected(session *stdtcp.Session, err error) {
	fmt.Println(session.GetMeta(), err)
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

	if msgName == (&models.PING{}).XXX_MessageName() {
		_ = session.Send(nil, &models.PONG{})
	} else {
		atomic.AddUint64(&count, 1)
		fmt.Println(msgcreate.MessageName(msg), msg)
		data := msg.(*models.Resp)
		for _, other := range data.OtherMsg {
			if other.TypeUrl == (&rlpb.RoguelikeMatch_RLRoguelikeStartPush{}).XXX_MessageName() {
				otherData := msgcreate.NewMessage(other.TypeUrl)
				err1 := proto.Unmarshal(other.Value, otherData)
				if err1 != nil {
					panic(err)
				}
				go func() {
					time.Sleep(5 * time.Second)
					resp := &servicepb.GameBattle_CPPEnterBattleResponse{}
					d := otherData.(*rlpb.RoguelikeMatch_RLRoguelikeStartPush)
					err = session.RPCRequestOut(nil, &servicepb.GameBattle_CPPEnterBattleRequest{Pos: &cppbattle.CPPBattle_Vec2{X: 0, Y: 0}, MapId: d.MapId, BattleServerId: d.BattleId, IsSingle: true, ConfigId: 2001}, resp)
					if err != nil {
						fmt.Println("GameBattle_CPPEnterBattleRequest painc")
						panic(err)
						os.Exit(1)
					}
					this_.battleClient = &Client{log: this_.log, type_: 1, roleId: this_.roleId}
					stdtcp.Connect(resp.Ip+":"+strconv.Itoa(int(resp.Port)), time.Second*3, true, this_.battleClient, this_.log, false)
				}()
			}
		}
	}
}

func (this_ *Client) Close() {
	this_.sess.Close(nil)
}

func main() {
	rand.Seed(time.Now().UnixNano())
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
	for i := 0; i < 220; i++ {
		c := NewClient("10.23.50.229:8071", log)
		m[c] = struct{}{}
		time.Sleep(time.Millisecond)
	}
	for {
		time.Sleep(time.Second)
	}
}
