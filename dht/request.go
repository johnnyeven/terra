package dht

import (
	"net"
)

type Request struct {
	remoteAddr net.Addr
	cmd        string
	ClientID   interface{}
	Data       interface{}
}
