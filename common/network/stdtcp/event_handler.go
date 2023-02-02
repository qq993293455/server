package stdtcp

type dispatch interface {
	OnConnected(session *Session)
	OnDisconnected(session *Session, err error)
	OnMessage(session *Session, msgName string, frame []byte)
	OnRequest(session *Session, rpcIndex uint32, msgName string, frame []byte)
}

type EventHandler struct {
}

// OnConnected 当监听连接建立时,当连接服务器成功时 会调用
func (this_ *EventHandler) OnConnected(session *Session) {

}

func (this_ *EventHandler) OnDisconnected(session *Session, err error) {

}

func (this_ *EventHandler) OnMessage(session *Session, frame []byte) {

}
