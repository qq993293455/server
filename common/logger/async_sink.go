package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"coin-server/common/values/env"

	"github.com/pkg/errors"
	"gopkg.in/natefinch/lumberjack.v2"
)

const AsyncSinkName = "AsyncSink"

type AsyncSink struct {
	tcpAddrS  []string
	filesPath []string
	console   string
	writers   []io.WriteCloser
	open      int32

	swapBuff [2][][]byte
	cond     *sync.Cond

	close_ uint32
	closed chan struct{}
}

func NewAsyncSink(tcpAddrS []string, files []string, console string) *AsyncSink {
	as := &AsyncSink{
		tcpAddrS:  tcpAddrS,
		filesPath: files,
		console:   console,
		writers:   nil,
		open:      0,
		swapBuff:  [2][][]byte{},
		cond:      sync.NewCond(&sync.Mutex{}),
	}
	writers, err := as.openWriter()
	if err != nil {
		panic(err)
	}
	as.writers = writers
	as.run()
	return as
}

const buffLen = 40960

var buffPool = sync.Pool{New: func() interface{} {
	return make([]byte, 0, buffLen)
}}

func getBuff() []byte {
	return buffPool.Get().([]byte)
}

func putBuff(b []byte) {
	if cap(b) > buffLen {
		return
	}
	b = b[:0]
	buffPool.Put(b)
}

func (this_ *AsyncSink) close() {
	atomic.StoreUint32(&this_.close_, 1)
}

func (this_ *AsyncSink) Closed() bool {
	return atomic.LoadUint32(&this_.close_) == 1
}

func (this_ *AsyncSink) openWriter() (writers []io.WriteCloser, err error) {
	defer func() {
		if err != nil {
			for _, v := range writers {
				_ = v.Close()
			}
		}
	}()
	if len(this_.tcpAddrS) > 0 {
		for _, v := range this_.tcpAddrS {
			var tcpW *TcpWriter
			tcpW, err = NewTcpWriter(v)
			if err != nil {
				return
			}
			writers = append(writers, tcpW)
		}
	}

	if len(this_.filesPath) > 0 {
		for _, v := range this_.filesPath {
			l := &lumberjack.Logger{
				Filename:   v,
				MaxSize:    int(env.GetLogMaxSize()), // megabytes
				MaxBackups: int(env.GetLogMaxBackUp()),
				MaxAge:     3650, //days
			}
			writers = append(writers, l)
		}
	}
	switch this_.console {
	case "stdout":
		writers = append(writers, os.Stdout)
	case "stderr":
		writers = append(writers, os.Stderr)
	}
	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}
	return
}

func (this_ *AsyncSink) Sync() error {
	return nil
}

func (this_ *AsyncSink) Close() error {
	this_.cond.Broadcast()
	this_.close()
	<-this_.closed
	var err error
	for _, v := range this_.writers {
		er := v.Close()
		if er != nil {
			err = errors.Wrap(err, er.Error())
		}
	}
	return err
}

func (this_ *AsyncSink) run() {
	go func() {
		defer func() {
			if e := recover(); e != nil {
				_, _ = fmt.Fprintln(os.Stderr, "panic:", e)
				this_.run()
			}
		}()
		for !this_.Closed() {
			this_.cond.L.Lock()
			this_.swapBuff[0], this_.swapBuff[1] = this_.swapBuff[1], this_.swapBuff[0]
			for len(this_.swapBuff[1]) == 0 && !this_.Closed() {
				this_.cond.Wait()
				this_.swapBuff[0], this_.swapBuff[1] = this_.swapBuff[1], this_.swapBuff[0]
			}
			this_.cond.L.Unlock()

			for _, v := range this_.swapBuff[1] {
				for _, writer := range this_.writers {
					_, err := writer.Write(v)
					if err != nil {
						_, _ = fmt.Fprintln(os.Stderr, err)
					}
				}
				putBuff(v)
			}
			this_.swapBuff[1] = this_.swapBuff[1][:0]
		}
		for i, buff := range this_.swapBuff {
			for _, v := range buff {
				for _, writer := range this_.writers {
					_, err := writer.Write(v)
					if err != nil {
						_, _ = fmt.Fprintln(os.Stderr, err)
					}
				}
				putBuff(v)
			}
			this_.swapBuff[i] = this_.swapBuff[i][:0]
		}
		close(this_.closed)
	}()
}

func (this_ *AsyncSink) Write(p []byte) (n int, err error) {
	lp := len(p)
	this_.cond.L.Lock()
	defer func() {
		this_.cond.L.Unlock()
		this_.cond.Signal()
	}()
	temp := this_.swapBuff[0]
	l := len(temp)
	if l > 0 {
		last := temp[l-1]
		if cap(last)-len(last) > lp {
			temp[l-1] = append(temp[l-1], p...)
			return lp, nil
		}
	}
	if lp > buffLen {
		this_.swapBuff[0] = append(this_.swapBuff[0], p)
		return lp, nil
	}
	b := getBuff()
	b = append(b, p...)
	this_.swapBuff[0] = append(this_.swapBuff[0], b)
	return lp, nil
}
