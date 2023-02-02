package player

import "coin-server/common/network/stdtcp"

type Connection struct {
	Sess                 *stdtcp.Session
	HandleOnConnected    func(session *stdtcp.Session)
	HandleOnDisconnected func(session *stdtcp.Session, err error)
	HandleOnMessage      func(session *stdtcp.Session, msgName string, frame []byte)
	HandleOnRequest      func(session *stdtcp.Session, rpcIndex uint32, msgName string, frame []byte)
}

func (this_ *Connection) OnConnected(session *stdtcp.Session) {
	if this_.HandleOnConnected != nil {
		this_.Sess = session
		this_.HandleOnConnected(session)
	}
}
func (this_ *Connection) OnDisconnected(session *stdtcp.Session, err error) {
	if this_.HandleOnDisconnected != nil {
		this_.Sess = nil
		this_.HandleOnDisconnected(session, err)
	}
}
func (this_ *Connection) OnMessage(session *stdtcp.Session, msgName string, frame []byte) {
	if this_.HandleOnMessage != nil {
		this_.HandleOnMessage(session, msgName, frame)
	}
}
func (this_ *Connection) OnRequest(session *stdtcp.Session, rpcIndex uint32, msgName string, frame []byte) {
	if this_.HandleOnRequest != nil {
		this_.Sess = nil
		this_.HandleOnRequest(session, rpcIndex, msgName, frame)
	}
}
