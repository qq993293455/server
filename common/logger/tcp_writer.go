package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"coin-server/common/utils"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type TcpWriter struct {
	addr      string
	conn      net.Conn
	isConnect int32
	tempFile  *os.File
}

func NewTcpWriter(addr string) (*TcpWriter, error) {
	tw := &TcpWriter{
		addr:      addr,
		conn:      nil,
		isConnect: 0,
		tempFile:  nil,
	}
	err := tw.connect_()
	if err != nil {
		return nil, err
	}
	exists := utils.CheckPathExists(addrToFileName(addr))
	if exists {
		tw.tempFile, err = os.OpenFile(addrToFileName(addr), os.O_RDONLY, 0666)
		if err != nil {
			return nil, err
		}
		err, err1 := tw.syncTempFile()
		if err != nil {
			return nil, err
		}
		if err1 != nil {
			return nil, err1
		}
	}
	return tw, nil
}

func addrToFileName(addr string) string {
	return "./tcp" + strings.ReplaceAll(addr, ":", "_") + ".fail.log"
}

func addrToFileIndexName(addr string) string {
	return addrToFileName(addr) + ".index"
}

func (this_ *TcpWriter) connect() {
	go func() {
		err := this_.connect_()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "tcpSink:", err)
			this_.connect()
		}
	}()
}

func (this_ *TcpWriter) connect_() error {
	conn, err := net.DialTimeout("tcp", this_.addr, time.Second*3)
	if err != nil {
		return err
	}
	this_.conn = conn
	atomic.StoreInt32(&this_.isConnect, 1)
	return nil
}

func (this_ *TcpWriter) Close() error {
	return nil
}

func (this_ *TcpWriter) syncTempFile() (fileErr error, tcpErr error) {
	if this_.tempFile == nil {
		return
	}
	fileErr = this_.tempFile.Sync()
	if fileErr != nil {
		return
	}
	indexFileName := addrToFileIndexName(this_.addr)
	index := int64(0)
	indexData, err := ioutil.ReadFile(indexFileName)
	if err != nil {
		if !os.IsNotExist(err) {
			fileErr = err
			return
		}
	}
	if len(indexData) != 0 {
		index, _ = strconv.ParseInt(string(indexData), 10, 64)
	}
	data := make([]byte, 10240)
	for {
		index, fileErr = this_.tempFile.Seek(index, 0)
		if fileErr != nil && fileErr != io.EOF {
			return
		}
		var readN int
		readN, fileErr = this_.tempFile.Read(data)
		if fileErr != nil {
			return
		}
		if readN > 0 {
			_, tcpErr = this_.Write(data[:readN])
			if tcpErr != nil {
				return
			}
		}
		index += int64(readN)
		if fileErr == io.EOF {
			_ = this_.tempFile.Close()
			_ = os.Remove(indexFileName)
			_ = os.Remove(addrToFileName(this_.addr))
			fileErr = nil
			return
		} else {
			_ = ioutil.WriteFile(indexFileName, []byte(strconv.Itoa(int(index))), 0666)
		}
	}
}

func (this_ *TcpWriter) CloseAndConnect() {
	atomic.StoreInt32(&this_.isConnect, 0)
	_ = this_.conn.Close()
	this_.connect()
}

func (this_ *TcpWriter) Write(b []byte) (n int, err error) {
	if atomic.LoadInt32(&this_.isConnect) == 1 {
		if this_.tempFile != nil {
			fileErr, tcpErr := this_.syncTempFile()
			if tcpErr != nil {
				return 0, err
			}
			if fileErr == nil {
				_, _ = fmt.Fprintln(os.Stderr, "tcpSink:fileErr", fileErr)
			}
		}
		n, err = this_.write(b)
		if err == nil {
			return
		}
		this_.CloseAndConnect()
		b = b[n:]
	}
	if this_.tempFile == nil {
		this_.tempFile, err = os.OpenFile(addrToFileName(this_.addr), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return 0, err
		}
	}

	return this_.tempFile.Write(b)
}

func (this_ *TcpWriter) write(b []byte) (n int, err error) {
	err = this_.conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
	if err != nil {
		return 0, err
	}
	defer func() {
		err = this_.conn.SetWriteDeadline(time.Time{})
	}()
	return this_.conn.Write(b)
}
