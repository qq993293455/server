package stdtcp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"coin-server/common/bytespool"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/protocol"
	"coin-server/common/safego"
	"coin-server/common/timer"

	"github.com/gogo/protobuf/proto"

	"go.uber.org/zap"
)

const (
	TotalLenLen = 4
)

type Session struct {
	rpcIndex uint32
	closed   uint32
	conn     net.Conn
	rAddr    string
	lAddr    string
	rIP      string

	sendBuf [2][][]byte
	cond    *sync.Cond

	once sync.Once

	log              *logger.Logger
	lastReadTime     int64
	isDebugMod       bool
	isClient         bool
	isReconnect      bool
	isAbortReconnect bool
	dailTimeout      time.Duration

	h dispatch
	// 是否已经验证是合法连接
	isCheck uint32

	version byte

	meta interface{}

	isForClient bool

	rpcMap   map[uint32]*rpcInfo
	rpcMutex sync.Mutex
}

func newErrClientSession(addr string, log *logger.Logger, isDebug bool, h dispatch, isReconnect bool, timeout time.Duration) *Session {
	s := &Session{
		rpcIndex:    0,
		closed:      0,
		conn:        nil,
		rAddr:       addr,
		lAddr:       "",
		sendBuf:     [2][][]byte{},
		cond:        sync.NewCond(&sync.Mutex{}),
		once:        sync.Once{},
		log:         log,
		isDebugMod:  isDebug,
		isClient:    true,
		isReconnect: isReconnect,
		h:           h,
		dailTimeout: timeout,
		rpcMap:      map[uint32]*rpcInfo{},
	}
	s.lastReadTime = timer.Now().Unix()
	return s
}

func newClientSession(conn net.Conn, log *logger.Logger, isDebug bool, h dispatch, isReconnect bool, timeout time.Duration) *Session {
	s := &Session{
		rpcIndex:    0,
		closed:      0,
		conn:        conn,
		rAddr:       "",
		lAddr:       "",
		sendBuf:     [2][][]byte{},
		cond:        sync.NewCond(&sync.Mutex{}),
		once:        sync.Once{},
		log:         log,
		isDebugMod:  isDebug,
		isClient:    true,
		isReconnect: isReconnect,
		h:           h,
		dailTimeout: timeout,
		rpcMap:      map[uint32]*rpcInfo{},
	}
	s.lastReadTime = timer.Now().Unix()
	s.rAddr = conn.RemoteAddr().String()
	s.lAddr = conn.LocalAddr().String()
	s.rIP = conn.RemoteAddr().(*net.TCPAddr).IP.String()
	return s
}

var checkData = []byte{0, 0xdd, 0, 0, 0, 0, 0, 0}
var checkLen = len(checkData)
var PINGPONGTimeout = errmsg.NewProtocolErrorInfo("ping-pong timeout")

func (this_ *Session) sendCheck() error {
	_, err := this_.conn.Write(checkData)
	return err
}

func newServerSession(conn net.Conn, log *logger.Logger, isDebug bool, h dispatch, isForClient bool) *Session {
	s := &Session{
		rpcIndex:    0,
		closed:      0,
		conn:        conn,
		rAddr:       "",
		lAddr:       "",
		sendBuf:     [2][][]byte{},
		cond:        sync.NewCond(&sync.Mutex{}),
		once:        sync.Once{},
		log:         log,
		isDebugMod:  isDebug,
		isClient:    false,
		h:           h,
		isForClient: isForClient,
		rpcMap:      map[uint32]*rpcInfo{},
	}
	s.lastReadTime = timer.Now().Unix()
	s.rAddr = conn.RemoteAddr().String()
	s.lAddr = conn.LocalAddr().String()
	s.rIP = conn.RemoteAddr().(*net.TCPAddr).IP.String()
	return s
}

func (this_ *Session) SetAbortReconnect() {
	this_.isAbortReconnect = true
}

func (this_ *Session) SetMeta(meta interface{}) {
	this_.meta = meta
}

func (this_ *Session) GetMeta() interface{} {
	return this_.meta
}

func (this_ *Session) LastReadTime() time.Time {
	return time.Unix(atomic.LoadInt64(&this_.lastReadTime), 0)
}

func (this_ *Session) RemoteAddr() string {
	return this_.rAddr
}

func (this_ *Session) LocalAddr() string {
	return this_.lAddr
}

func (this_ *Session) RemoteIP() string {
	return this_.rIP
}

func (this_ *Session) IsClose() bool {
	return atomic.LoadUint32(&this_.closed) == 1
}

func (this_ *Session) IsClient() bool {
	return this_.isClient
}

// Close 如果是connector并且开启了reconnect. ，当err==nil时，关闭并且停止重连
func (this_ *Session) Close(err error) {
	this_.once.Do(func() {
		if this_.conn != nil {
			_ = this_.conn.Close()
		}
		atomic.StoreUint32(&this_.closed, 1)
		this_.cond.Broadcast()
		this_.log.Debug("session closed", zap.Error(err), zap.String("remoteAddr", this_.rAddr))
		this_.h.OnDisconnected(this_, err)
		if err != nil {
			if this_.isClient && this_.isReconnect && !this_.isAbortReconnect {
				if this_.isDebugMod {
					this_.log.Debug("reconnecting......", zap.String("remoteAddr", this_.rAddr))
				}
				Connect(this_.rAddr, this_.dailTimeout, this_.isReconnect, this_.h, this_.log, this_.isDebugMod)
			}
		}
	})
}

func (this_ *Session) Send(h *models.ServerHeader, msg proto.Message) *errmsg.ErrMsg {
	s := protocol.GetEncodeLen(h, msg)
	d := bytespool.GetSample(s)
	err := protocol.EncodeTCPFrom(d, uint8(models.MsgType_push), 0, h, msg)
	if err != nil {
		return err
	}
	return this_.write(d)
}

var ErrMaxSendDataLen = fmt.Errorf("Tcp Write error: data too long: %d ", maxSendDataLen)

// WriteSized 发送封包好的数据出去
func (this_ *Session) Write(data []byte) *errmsg.ErrMsg {
	return this_.write(data)
}

var errorSessionClosed = errmsg.NewProtocolErrorInfo("session closed")

func (this_ *Session) write(data []byte) *errmsg.ErrMsg {
	if this_.IsClose() {
		return errorSessionClosed
	}
	this_.cond.L.Lock()
	this_.sendBuf[0] = append(this_.sendBuf[0], data)
	this_.cond.L.Unlock()
	this_.cond.Signal()
	return nil
}

const (
	maxSendDataLen = int(1 << 24)
	maxMsgLen      = maxSendDataLen
)

func (this_ *Session) writeLoop() {
	safego.GOWithLogger(this_.log, func() {
		var err error
		defer func() {
			if this_.isDebugMod {
				this_.log.Debug("write closed", zap.String("remoteAddr", this_.rAddr), zap.String("localAddr", this_.lAddr), zap.Error(err))
			}
			this_.Close(err)
		}()
		//if this_.isClient {
		//	err = this_.sendCheck()
		//	if err != nil {
		//		return
		//	}
		//}

		for !this_.IsClose() {
			this_.cond.L.Lock()
			this_.sendBuf[0], this_.sendBuf[1] = this_.sendBuf[1], this_.sendBuf[0]
			for len(this_.sendBuf[1]) == 0 {
				this_.cond.Wait()
				if this_.IsClose() {
					return
				}
				this_.sendBuf[0], this_.sendBuf[1] = this_.sendBuf[1], this_.sendBuf[0]
			}
			this_.cond.L.Unlock()
			temp := this_.sendBuf[1]
			for len(temp) > 0 {
				size, index := getMaxSendDataLen(temp)
				if index == 1 {
					v := temp[0]
					_, err = this_.conn.Write(v)
					bytespool.PutSample(v)
					if err != nil {
						return
					}
				} else {
					d := bytespool.GetSample(size)
					s := 0
					for x := 0; x < index; x++ {
						s += copy(d[s:], temp[x])
						bytespool.PutSample(temp[x])
					}
					_, err = this_.conn.Write(d) // 批量写。以达到最少的IO次数
					bytespool.PutSample(d)
					if err != nil {
						return
					}
				}
				temp = temp[index:]
			}
			this_.sendBuf[1] = this_.sendBuf[1][:0]
		}
	})
}

const idealSendLen = 16384

func getMaxSendDataLen(d [][]byte) (int, int) {
	size := 0
	for i, v := range d {
		lv := len(v)
		if size+lv > idealSendLen {
			if size == 0 {
				return size + lv, i + 1
			}
			return size, i + 1
		}
		size += len(v)
	}
	return size, len(d)
}

func (this_ *Session) start() {
	this_.log.Debug("connect success", zap.String("remoteAddr", this_.rAddr), zap.String("localAddr", this_.lAddr))
	if !this_.isClient && this_.isForClient {
		this_.readLoopForClient()
	} else {
		this_.readLoop()
	}
	this_.writeLoop()
}

func (this_ *Session) readLoopForClient() {
	safego.GOWithLogger(this_.log, func() {
		var err error
		defer func() {
			if this_.isDebugMod {
				this_.log.Debug("read closed", zap.String("remoteAddr", this_.rAddr), zap.String("localAddr", this_.lAddr), zap.Error(err))
			}
			this_.Close(err)
		}()

		header := make([]byte, TotalLenLen)
		var frame []byte
		for !this_.IsClose() {
			frame, err = this_.readOneForClient(header)
			if err != nil {
				return
			}
			atomic.StoreInt64(&this_.lastReadTime, timer.Now().Unix())
			this_.OnMessage(frame)
			bytespool.PutSample(frame)
		}
	})
}

var errReadHeaderFailed = errors.New("read header failed")

func (this_ *Session) readOneForClient(header []byte) ([]byte, error) {
	n, err := this_.conn.Read(header)
	if err != nil {
		return nil, err
	}
	if n != len(header) {
		return nil, errReadHeaderFailed
	}
	totalLen := int(binary.LittleEndian.Uint32(header))
	if totalLen > maxMsgLen {
		return nil, fmt.Errorf("tcp recv data : msgLen too big %d > max:%d ", totalLen, maxMsgLen)
	}
	if totalLen < TotalLenLen {
		return nil, fmt.Errorf("msgLen too small %d ", totalLen)
	}
	data := bytespool.GetSample(totalLen)
	copy(data, header)
	start := TotalLenLen
	for {
		n, err := this_.conn.Read(data[start:])
		if err != nil {
			return nil, err
		}
		start += n
		if start == totalLen {
			break
		}
	}
	return data, nil
}

func (this_ *Session) readLoop() {
	safego.GOWithLogger(this_.log, func() {
		var err error
		defer func() {
			if this_.isDebugMod {
				this_.log.Debug("read closed", zap.String("remoteAddr", this_.rAddr), zap.String("localAddr", this_.lAddr), zap.Error(err))
			}
			this_.Close(err)
		}()
		const onceReadLen = 16384 / 2 / 2
		readData := make([]byte, onceReadLen)
		readStart := 0
		var n int

		for !this_.IsClose() {
			n, err = this_.conn.Read(readData[readStart:])

			if err != nil {
				return
			}
			readStart += n
			if readStart < TotalLenLen { // 不满totalLen个字节，继续读
				continue
			}
			gotData := readData[:readStart]
			for {
				if len(gotData) < TotalLenLen { // 比总长度小，继续读
					break
				}
				msgLen := int(binary.LittleEndian.Uint32(gotData))
				if msgLen > maxMsgLen { // 大于最大长度，直接关闭
					err = fmt.Errorf("msgLen too big %d > max:%d", msgLen, maxMsgLen)
					return
				}
				if msgLen < TotalLenLen { // 大于最小长度，直接关闭
					err = fmt.Errorf("msgLen too small %d", msgLen)
					return
				}
				var tempData []byte
				if msgLen > len(gotData) {
					if msgLen > onceReadLen {
						tempData = make([]byte, msgLen)
						copy(tempData, gotData)
						_, err = io.ReadFull(this_.conn, tempData[len(gotData):])

						if err != nil {
							return
						}
						atomic.StoreInt64(&this_.lastReadTime, timer.Now().Unix())
						this_.OnMessage(tempData)
						gotData = nil
					} else {
						break
					}
				} else {
					atomic.StoreInt64(&this_.lastReadTime, timer.Now().Unix())
					this_.OnMessage(gotData[:msgLen])
					gotData = gotData[msgLen:]
				}
			}
			lg := len(gotData)
			if lg > 0 {
				copy(readData, gotData)
				readStart = lg
			} else {
				readStart = 0
			}
		}
	})
}

func (this_ *Session) OnMessage(frame []byte) {
	defer safego.RecoverWithLogger(this_.log)

	msgType, rpcIndex, msgName, raw, err := protocol.DecodeTCPRaw(frame)
	if err != nil {
		this_.Close(err)
		return
	}
	now := timer.Now()
	defer func() {
		cost := timer.Now().Sub(now)
		if cost > time.Second {
			this_.log.Debug("OnMessage end", zap.String("remoteAddr", this_.rAddr), zap.String("localAddr", this_.lAddr), zap.String("msgName", msgName), zap.Duration("cost", cost))
		}
	}()
	switch models.MsgType(msgType) {
	case models.MsgType_response:
		this_.rpcMutex.Lock()
		ri, ok := this_.rpcMap[rpcIndex]
		delete(this_.rpcMap, rpcIndex)
		this_.rpcMutex.Unlock()
		if ok && atomic.CompareAndSwapUint32(&ri.done, 0, 1) {
			resp := &models.Resp{}
			err := protocol.DecodeInternal(raw, nil, resp)
			if err != nil {
				if ri.callback != nil {
					ri.callback(nil, err)
				}
			} else {
				if ri.callback != nil {
					ri.callback(resp, nil)
				}
			}
		}
	case models.MsgType_push:
		this_.h.OnMessage(this_, msgName, raw)
	case models.MsgType_request:
		this_.h.OnRequest(this_, rpcIndex, msgName, raw)
	}
}

func (this_ *Session) RPCResponse(rpcIndex uint32, h *models.ServerHeader, msg proto.Message, otherMsgS ...proto.Message) *errmsg.ErrMsg {
	resp := &models.Resp{
		ErrCode:         0,
		ErrMsg:          "",
		ErrInternalInfo: "",
		Resp:            protocol.MsgToAny(msg),
		OtherMsg:        nil,
		OtherRequest:    nil,
	}
	for _, v := range otherMsgS {
		resp.OtherMsg = append(resp.OtherMsg, protocol.MsgToAny(v))
	}
	n := protocol.GetEncodeLen(h, resp)
	d := bytespool.GetSample(n)
	err := protocol.EncodeTCPFrom(d, uint8(models.MsgType_response), rpcIndex, h, resp)
	if err != nil {
		return err
	}
	return this_.write(d)
}

func (this_ *Session) RPCResponseTCP(rpcIndex uint32, msgName string, respData []byte) *errmsg.ErrMsg {
	n := protocol.GetEncodeInternalDataLen(msgName, respData)
	d := bytespool.GetSample(n)
	err := protocol.EncodeTCPInternalDataFrom(d, uint8(models.MsgType_response), rpcIndex, msgName, respData)
	if err != nil {
		return err
	}
	return this_.write(d)
}

func (this_ *Session) RPCResponseError(rpcIndex uint32, h *models.ServerHeader, err *errmsg.ErrMsg) *errmsg.ErrMsg {
	if err == nil {
		panic("err is nil")
	}
	resp := (*models.Resp)(err)
	n := protocol.GetEncodeLen(h, resp)
	d := bytespool.GetSample(n)
	err = protocol.EncodeTCPFrom(d, uint8(models.MsgType_response), rpcIndex, h, resp)
	if err != nil {
		return err
	}
	return this_.write(d)
}

var chanPool = sync.Pool{New: func() interface{} {
	return make(chan struct{})
}}

func (this_ *Session) RPCRequest(h *models.ServerHeader, msg proto.Message, out **models.Resp) *errmsg.ErrMsg {
	var respErr *errmsg.ErrMsg
	ci := chanPool.Get()
	defer chanPool.Put(ci)
	c := ci.(chan struct{})
	this_.AsyncRPCRequest(h, msg, func(resp *models.Resp, err *errmsg.ErrMsg) {
		*out = resp
		respErr = err
		c <- struct{}{}
	})
	<-c
	return respErr
}

func (this_ *Session) RPCRequestOut(h *models.ServerHeader, msg proto.Message, outMsg proto.Message) *errmsg.ErrMsg {
	var respErr *errmsg.ErrMsg
	ci := chanPool.Get()
	defer chanPool.Put(ci)
	c := ci.(chan struct{})
	var out *models.Resp
	this_.AsyncRPCRequest(h, msg, func(resp *models.Resp, err *errmsg.ErrMsg) {
		out = resp
		respErr = err
		c <- struct{}{}
	})
	<-c
	if respErr != nil {
		return respErr
	}
	if out.ErrCode != models.ErrorType_NoError {
		return (*errmsg.ErrMsg)(out)
	}
	if out.Resp == nil {
		return nil
	}

	e := proto.Unmarshal(out.Resp.Value, outMsg)
	if e != nil {
		return errmsg.NewProtocolError(e)
	}
	return nil
}

type rpcInfo struct {
	rpcIndex uint32
	timeout  time.Time
	callback Callback
	done     uint32
}

type Callback func(message *models.Resp, err *errmsg.ErrMsg)

var errTimeout = errmsg.NewProtocolErrorInfo("rpc timeout")

func (this_ *Session) AsyncRPCRequest(h *models.ServerHeader, msg proto.Message, callback Callback) {
	rpcIndex := atomic.AddUint32(&this_.rpcIndex, 1)
	s := protocol.GetEncodeLen(h, msg)
	d := bytespool.GetSample(s)

	err := protocol.EncodeTCPFrom(d, uint8(models.MsgType_request), rpcIndex, h, msg)
	if err != nil {
		if callback != nil {
			callback(nil, err)
		}
		return
	}
	timeout := time.Second * 10
	ri := &rpcInfo{
		rpcIndex: rpcIndex,
		timeout:  timer.Now().Add(timeout),
		callback: callback,
		done:     0,
	}
	this_.rpcMutex.Lock()
	this_.rpcMap[rpcIndex] = ri
	this_.rpcMutex.Unlock()
	timer.AfterFunc(timeout, func() {
		if atomic.CompareAndSwapUint32(&ri.done, 0, 1) {
			if callback != nil {
				callback(nil, errTimeout)
			}
			this_.rpcMutex.Lock()
			delete(this_.rpcMap, rpcIndex)
			this_.rpcMutex.Unlock()
		}
	})
	err = this_.Write(d)
	if err != nil {
		if atomic.CompareAndSwapUint32(&ri.done, 0, 1) {
			if callback != nil {
				callback(nil, err)
			}
			this_.rpcMutex.Lock()
			delete(this_.rpcMap, rpcIndex)
			this_.rpcMutex.Unlock()
		}
		return
	}
}
