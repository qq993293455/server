package stdtcp

import (
	"net"
	"sync"
	"time"

	"coin-server/common/gopool"
	"coin-server/common/logger"
	"coin-server/common/reuseport"
	"coin-server/common/safego"

	"go.uber.org/zap"
)

type Acceptor struct {
	h           dispatch
	addr        string
	close       chan struct{}
	listener    net.Listener
	log         *logger.Logger
	isDebugMod  bool
	once        sync.Once
	isForClient bool
}

func NewAcceptor(addr string, log *logger.Logger, isDebug bool, h dispatch, isForClient bool) (*Acceptor, error) {
	listener, err := reuseport.Listen("tcp4", addr)
	if err != nil {
		return nil, err
	}
	a := &Acceptor{
		h:           h,
		addr:        addr,
		close:       make(chan struct{}),
		listener:    listener,
		log:         log,
		isDebugMod:  isDebug,
		once:        sync.Once{},
		isForClient: isForClient,
	}
	return a, nil
}

func (this_ *Acceptor) Start() {
	safego.GO(func(i interface{}) {
		this_.log.Error("acceptor listen panic", zap.Any("panic info", i))
		this_.Start()
	}, func() {
		var tempDelay time.Duration // 监听失败时暂停多久重新开始接收
		this_.log.Info("acceptor start listen ......", zap.String("addr", this_.addr))
		for {
			conn, err := this_.listener.Accept()
			if err != nil {
				select {
				case <-this_.close:
					return
				default:
				}
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					if tempDelay == 0 {
						tempDelay = 5 * time.Millisecond
					} else {
						tempDelay *= 2
					}
					if max := 1 * time.Second; tempDelay > max {
						tempDelay = max
					}
					this_.log.Warn("tcp: accept error", zap.Error(err), zap.Duration("tempDelay", tempDelay))
					time.Sleep(tempDelay)
					continue
				}
				return
			}
			this_.startConn(conn)
		}
	})
}

func (this_ *Acceptor) startConn(conn net.Conn) {
	gopool.Submit(func() {
		sess := newServerSession(conn, this_.log, this_.isDebugMod, this_.h, this_.isForClient)
		sess.start()
		this_.h.OnConnected(sess)
	})
}

func (this_ *Acceptor) Close() {
	close(this_.close)
	_ = this_.listener.Close()
}
