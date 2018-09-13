package dht

import "net"

type Packet struct {
	Data       []byte
	RemoteAddr *net.UDPAddr
}

type TransportDriver interface {
	net.Conn
	MakeRequest(requestType string, data map[string]interface{}, target net.Addr) *Request
	Request(request *Request)
	sendRequest(request *Request, retry int)
	Receive(receiveChannel chan Packet)
}
