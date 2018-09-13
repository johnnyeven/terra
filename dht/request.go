package dht

import (
	"net"
)

type Request struct {
	remoteAddr net.Addr
	cmd        string
	Data       map[string]interface{}
}
