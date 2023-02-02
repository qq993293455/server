package netstat

import (
	"fmt"
	"net"
)

// SockAddr represents an ip:port pair
type SockAddr struct {
	IP   net.IP
	Port uint16
}

func (s *SockAddr) String() string {
	return fmt.Sprintf("%v:%d", s.IP, s.Port)
}

// SockTabEntry type represents each line of the /proc/net/[tcp|udp]
type SockTabEntry struct {
	ino        string
	LocalAddr  *SockAddr
	RemoteAddr *SockAddr
	State      SkState
	UID        uint32
	Process    *Process
}

// Process holds the PID and process name to which each socket belongs
type Process struct {
	Pid  int
	Name string
}

func (p *Process) String() string {
	return fmt.Sprintf("%d/%s", p.Pid, p.Name)
}

// SkState type represents socket connection state
type SkState uint8

func (s SkState) String() string {
	return skStates[s]
}

// AcceptFn is used to filter socket entries. The value returned indicates
// whether the element is to be appended to the socket list.
type AcceptFn func(*SockTabEntry) bool

// NoopFilter - a test function returning true for all elements
func NoopFilter(*SockTabEntry) bool { return true }

// TCPSocks returns a slice of active TCP sockets containing only those
// elements that satisfy the accept function
func TCPSocks(accept AcceptFn) ([]SockTabEntry, error) {
	return osTCPSocks(accept)
}

// TCP6Socks returns a slice of active TCP IPv4 sockets containing only those
// elements that satisfy the accept function
func TCP6Socks(accept AcceptFn) ([]SockTabEntry, error) {
	return osTCP6Socks(accept)
}

// UDPSocks returns a slice of active UDP sockets containing only those
// elements that satisfy the accept function
func UDPSocks(accept AcceptFn) ([]SockTabEntry, error) {
	return osUDPSocks(accept)
}

// UDP6Socks returns a slice of active UDP IPv6 sockets containing only those
// elements that satisfy the accept function
func UDP6Socks(accept AcceptFn) ([]SockTabEntry, error) {
	return osUDP6Socks(accept)
}

func UsedPorts(start, end uint16, selfIPS ...string) (map[uint16]SkState, error) {
	if start > end {
		start, end = end, start
	}
	tcpSocks, err := TCPSocks(NoopFilter)
	if err != nil {
		return nil, err
	}
	tcp6Socks, err := TCP6Socks(NoopFilter)
	if err != nil {
		return nil, err
	}

	udpSocks, err := UDPSocks(NoopFilter)
	if err != nil {
		return nil, err
	}
	udp6Socks, err := UDP6Socks(NoopFilter)
	if err != nil {
		return nil, err
	}
	out := make(map[uint16]SkState, 1024)
	checkPorts(out, tcpSocks, start, end, selfIPS...)
	checkPorts(out, tcp6Socks, start, end, selfIPS...)
	checkPorts(out, udpSocks, start, end, selfIPS...)
	checkPorts(out, udp6Socks, start, end, selfIPS...)
	return out, nil
}

func checkPorts(out map[uint16]SkState, entry []SockTabEntry, start, end uint16, selfIPS ...string) {
	for i := range entry {
		v := &entry[i]
		if v.LocalAddr != nil {
			lp := v.LocalAddr.Port
			if lp != 0 && lp >= start && lp <= end {
				out[lp] = v.State
			}
		}
		if v.RemoteAddr != nil {
			lp := v.RemoteAddr.Port
			if lp != 0 && lp >= start && lp <= end {
				out[lp] = v.State
			}
		}
	}
}
