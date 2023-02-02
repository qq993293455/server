package service

import (
	"coin-server/common/safego"
	"coin-server/new-center-server/service/edge"
	"net/netip"
	"strconv"
	"time"

	"coin-server/common/logger"
	"coin-server/common/network/stdtcp"
	"coin-server/common/proto/models"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"

	"go.uber.org/zap"
)

type Service struct {
	acceptor          *stdtcp.Acceptor
	log               *logger.Logger
	svc               *service.Service
	serverId          values.ServerId
	serverType        models.ServerType
	selfAddr          string
	initLinesFlag     int //挂机副本 0 未初始化分綫 1 正在初始化分綫 2 完成初始化分綫
	initBossHallFlag  int //恶魔秘境 0 未初始化分綫 1 正在初始化分綫 2 完成初始化分綫
	freeLinesMap      map[values.Integer]values.Integer
	freeBossHallLines map[values.Integer]values.Integer
}

func NewCenterService(log *logger.Logger, addr string, natsUrls []string, serverId values.ServerId) *Service {
	s := &Service{
		acceptor:          nil,
		log:               log,
		svc:               nil,
		serverId:          serverId,
		serverType:        models.ServerType_CenterServer,
		selfAddr:          addr,
		initLinesFlag:     0,
		initBossHallFlag:  0,
		freeLinesMap:      map[values.Integer]values.Integer{},
		freeBossHallLines: map[values.Integer]values.Integer{},
	}
	s.svc = service.NewService(natsUrls, log, s.serverId, s.serverType, true, true)
	s.Router()
	edge.InitEdge(log)
	//检测分线人数
	safego.GOWithLogger(log, func() {
		s.checkLineValid()
	})

	safego.GOWithLogger(log, func() {
		s.checkRoleValid()
	})

	time.AfterFunc(10*time.Second, func() {
		s.log.Debug("reach timer,start checkBattleLines")
		for s.initLinesFlag != 2 {
			if s.initLinesFlag == 0 {
				s.initLinesFlag = 1
			}
			s.checkBattleLines(true)
			if s.initLinesFlag == 2 {
				break
			}
			time.Sleep(1 * time.Second)
		}

	})

	time.AfterFunc(10*time.Second, func() {
		s.log.Debug("reach timer,start checkBossHalleLines")
		for s.initBossHallFlag != 2 {
			if s.initBossHallFlag == 0 {
				s.initBossHallFlag = 1
			}
			s.checkBossHallLines(true)
			if s.initBossHallFlag == 2 {
				break
			}
			time.Sleep(1 * time.Second)
		}

	})

	//检测恶魔秘境分线人数
	safego.GOWithLogger(log, func() {
		s.checkBossHallLineValid()
	})

	return s
}

func (this_ *Service) Run() error {
	ap, err := netip.ParseAddrPort(this_.selfAddr)
	if err != nil {
		return err
	}
	la := ":" + strconv.Itoa(int(ap.Port()))
	this_.acceptor, err = stdtcp.NewAcceptor(la, this_.log, true, this_, false)
	if err != nil {
		return err
	}
	this_.acceptor.Start()
	this_.svc.Start(func(i interface{}) {
		this_.log.Warn("unknown event", zap.Any("event", i))
	}, false)
	return nil
}

func (this_ *Service) Close() {
	this_.svc.Close()
	this_.acceptor.Close()
}

func (this_ *Service) OnConnected(session *stdtcp.Session) {
	this_.log.Info("connect success", zap.String("remote", session.RemoteAddr()))
	timer.Ticker(time.Second*2, func() bool {
		if session.IsClose() {
			return false
		}
		if timer.Now().Sub(session.LastReadTime()) > time.Second*4 {
			session.Close(stdtcp.PINGPONGTimeout)
			return false
		}
		return true
	})
}

func (this_ *Service) OnDisconnected(session *stdtcp.Session, err error) {
	this_.log.Error("disconnect success", zap.String("remote", session.RemoteAddr()), zap.Error(err))
	typ, err1 := edge.OnSessionClose(session)
	if err1 != nil {
		this_.log.Error("OnSessionClose err", zap.String("remote", session.RemoteAddr()), zap.Error(err1))
		return
	}
	if typ == models.EdgeType_StaticServer {
		this_.checkBattleLines(false)
	}
	if typ == models.EdgeType_DynamicServer {
		this_.checkBossHallLines(false)
	}
}

func (this_ *Service) OnRequest(session *stdtcp.Session, rpcIndex uint32, msgName string, frame []byte) {
	err := this_.svc.HandleTCPData(session, rpcIndex, msgName, frame)
	if err != nil {
		session.Close(err)
	}
}

func (this_ *Service) OnMessage(session *stdtcp.Session, msgName string, frame []byte) {
	if msgName == (&models.PING{}).XXX_MessageName() {
		_ = session.Send(nil, &models.PONG{})
		return
	}
	err := this_.svc.HandleTCPData(session, 0, msgName, frame)
	if err != nil {
		session.Close(err)
	}
}
