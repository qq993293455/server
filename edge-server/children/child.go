package children

import (
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"coin-server/common/logger"
	"coin-server/common/msgcreate"
	"coin-server/common/network/stdtcp"
	"coin-server/common/proto/cppbattle"
	"coin-server/common/proto/edge"
	"coin-server/common/proto/models"
	"coin-server/common/protocol"

	"go.uber.org/zap"
)

type PlayerInfo struct {
	AddTime int64
}

type WaitListenPort struct {
	BattleId int64
	mu       sync.Mutex
	isClose  bool
	PortChan chan uint16 `json:"-"`
}

func (this_ *WaitListenPort) Close(port uint16) {
	this_.mu.Lock()
	defer this_.mu.Unlock()
	if this_.isClose {
		return
	}
	this_.isClose = true
	this_.PortChan <- port
}

type Child struct {
	EdgeId            string
	NatsUrls          []string
	BattleId          int64 // 战斗服ID
	MapId             int64 // 地图ID
	Error             error
	Program           string
	Args              []string
	IP                string
	Port              uint16
	ProcAttr          os.ProcAttr
	Process           *os.Process
	ProcessState      *os.ProcessState
	OnStarted         func(*Child)                             `json:"-"`
	OnStart           func(*Child)                             `json:"-"`
	OnExit            func(*Child)                             `json:"-"`
	OnPlayerChange    func(*Child, *edge.Edge_PlayerChange)    `json:"-"`
	OnNotCanEnterPush func(*Child, *edge.Edge_NotCanEnterPush) `json:"-"`
	CpuWeight         int64
	CreateInfo        *edge.Edge_CreateServerRequest
	closed            int64
	Log               *logger.Logger `json:"-"`
	CannotEnter       bool

	PlayerInfo map[string]*PlayerInfo
	once       sync.Once

	Sess       atomic.Value    `json:"-"`
	Wlp        *WaitListenPort `json:"-"`
	centerKill uint32
}

func (this_ *Child) SetCenterKill() {
	atomic.StoreUint32(&this_.centerKill, 1)
}

func (this_ *Child) IsCenterKill() bool {
	return atomic.LoadUint32(&this_.centerKill) == 1
}

func (this_ *Child) createNatsUrls() string {
	var str []string
	for _, v := range this_.NatsUrls {
		str = append(str, fmt.Sprintf(`"%s"`, v))
	}
	return strings.Join(str, ",")
}

func (this_ *Child) start() error {
	_, err := net.LookupHost(this_.IP)
	if err != nil {
		return err
	}

	if this_.Port < 20000 || this_.Port > 65535 {
		return fmt.Errorf("invalid Port [%d]", this_.Port)
	}
	if this_.BattleId <= 0 {
		return fmt.Errorf("invalid BattleId [%d]", this_.BattleId)
	}
	if this_.MapId <= 0 {
		return fmt.Errorf("invalid MapId [%d]", this_.MapId)
	}
	program := filepath.Join(this_.ProcAttr.Dir, this_.Program)
	program, err = exec.LookPath(program)
	if err != nil {
		return err
	}
	params, err := json.Marshal(this_.CreateInfo)
	if err != nil {
		return err
	}
	_, programName := filepath.Split(program)
	this_.ProcAttr.Files = append(this_.ProcAttr.Files, os.Stdin, os.Stdout, os.Stderr)
	this_.Args = append([]string{
		program,
		"-n", this_.createNatsUrls(),
		"-s", strconv.Itoa(int(this_.BattleId)),
		"-m", strconv.Itoa(int(this_.MapId)),
		"-i", this_.IP,
		"-p", strconv.Itoa(int(this_.Port)),
		"-l", filepath.Join("log", programName+".log"),
		"-o", "0",
		"-a", "1",
		"-t", "1",
		"-v", "4",
		"-r", string(params),
		"-e", this_.EdgeId,
	}, this_.Args...)

	this_.Process, err = os.StartProcess(program, this_.Args, &this_.ProcAttr)
	return err
}

func (this_ *Child) StartAndWait() {
	this_.Sess.Store((*stdtcp.Session)(nil))
	if this_.PlayerInfo == nil {
		this_.PlayerInfo = map[string]*PlayerInfo{}
	}
	this_.Error = this_.start()
	if this_.Error != nil {
		this_.safeCallOnExit()
		this_.OnStarted(this_)
		return
	}
	port := <-this_.Wlp.PortChan
	if port == 0 {
		this_.Error = fmt.Errorf("wait port failed:%d", port)
		this_.safeCallOnExit()
		this_.OnStarted(this_)
		return
	}
	this_.Port = port
	this_.startTCP()
	this_.safeCallOnStart()
	this_.ProcessState, this_.Error = this_.Process.Wait()
	this_.safeCallOnExit()
}

func (this_ *Child) startTCP() {
	addr := netip.AddrPortFrom(netip.MustParseAddr("127.0.0.1"), this_.Port).String()
	stdtcp.Connect(addr, time.Second, true, this_, this_.Log, true)
}

func (this_ *Child) OnConnected(session *stdtcp.Session) {
	this_.Sess.Store(session)
	push := &cppbattle.NSNB_EdgeAuth{}
	err := session.Send(nil, push)
	if err != nil {
		this_.Log.Error("OnConnected Send cppbattle.NSNB_EdgeAuth error", zap.Error(err))
	}
}

func (this_ *Child) GetSession() (*stdtcp.Session, bool) {
	v := this_.Sess.Load()
	if v == nil {
		return nil, false
	}
	s, ok := v.(*stdtcp.Session)
	if !ok || s == nil {
		return nil, false
	}
	return s, true
}

func (this_ *Child) OnDisconnected(session *stdtcp.Session, err error) {
	time.Sleep(time.Second)
	this_.Sess.Store((*stdtcp.Session)(nil))
	if atomic.LoadInt64(&this_.closed) == 1 {
		session.SetAbortReconnect()
	}

}

func (this_ *Child) OnMessage(session *stdtcp.Session, msgName string, frame []byte) {
	header := &models.ServerHeader{}
	out := msgcreate.NewMessage(msgName)
	err := protocol.DecodeInternal(frame, header, out)
	if err != nil {
		this_.Log.Error("protocol.DecodeInternal error", zap.String("msgName", msgName))
		return
	}
	switch msg := out.(type) {
	case *edge.Edge_PlayerChange:
		this_.once.Do(func() {
			this_.OnStarted(this_)
		})
		if this_.OnPlayerChange != nil {
			this_.OnPlayerChange(this_, msg)
		}
	case *edge.Edge_NotifyEdgeKillSelfPush:
		e := this_.Process.Kill()
		if e != nil {
			this_.Log.Info("Edge_NotifyEdgeKillSelfPush", zap.Int64("battle_id", this_.BattleId), zap.String("data", msg.String()))
		}
	case *edge.Edge_NotCanEnterPush:
		this_.OnNotCanEnterPush(this_, msg)
	default:
		this_.Log.Warn("receive unknown msg", zap.String("msgName", msgName), zap.String("data", msg.String()))
	}
}

func (this_ *Child) OnRequest(session *stdtcp.Session, rpcIndex uint32, msgName string, frame []byte) {
	this_.Log.Error("invalid OnRequest", zap.String("msgName", msgName))
}

func (this_ *Child) safeCallOnExit() {
	atomic.StoreInt64(&this_.closed, 1)
	if this_.OnExit != nil {
		this_.OnExit(this_)
	}
}

func (this_ *Child) safeCallOnStart() {
	if this_.OnStart != nil {
		this_.OnStart(this_)
	}
}
