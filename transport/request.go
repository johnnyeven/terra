package transport

import "net"

type Request struct {
	remoteAddr net.Addr
	cmd string
	data map[string]interface{}
}
