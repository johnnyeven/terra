package dht

import (
	"net"
)

type Request struct {
	remoteAddr net.Addr
	cmd        string
	ClientID   *Identity
	Data       map[string]interface{}
}
