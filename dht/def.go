package dht

import "net"

type Packet struct {
	Data       []byte
	RemoteAddr *net.UDPAddr
}

type TransportDriver interface {
	net.Conn
	MakeRequest(id *Identity, remoteAddr net.Addr, requestType string, data map[string]interface{}) *Request
	MakeResponse(remoteAddr net.Addr, tranID string, data map[string]interface{}) *Request
	MakeError(remoteAddr net.Addr, tranID string, errCode int, errMsg string) *Request
	Request(request *Request)
	sendRequest(request *Request, retry int)
	Receive(receiveChannel chan Packet)
}
