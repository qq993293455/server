package service

import "coin-server/common/network/stdtcp"

type Disconnection struct {
	Session *stdtcp.Session
}
