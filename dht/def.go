package dht

import "net"

type Packet struct {
	Data       []byte
	RemoteAddr *net.UDPAddr
}

type TransportDriver interface {
	net.Conn
	MakeRequest(id interface{}, remoteAddr net.Addr, requestType string, data interface{}) *Request
	MakeResponse(remoteAddr net.Addr, tranID string, data interface{}) *Request
	MakeError(remoteAddr net.Addr, tranID string, errCode int, errMsg string) *Request
	Request(request *Request)
	SendRequest(request *Request, retry int)
	Receive(receiveChannel chan Packet)
}
