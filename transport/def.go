package transport

import "net"

type TransportDriver interface {
	net.Conn
	MakeRequest(requestType string, data map[string]interface{}, target net.Addr) *Request
	Request(request *Request) error
	Receive(receiveChannel chan Packet)
}
