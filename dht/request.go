package dht

import (
	"net"
)

type Request struct {
	RemoteAddr net.Addr
	cmd        string
	ClientID   interface{}
	Data       interface{}
}
