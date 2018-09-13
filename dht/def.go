package dht

import "net"

type Packet struct {
	Data       []byte
	RemoteAddr *net.UDPAddr
}

type TransportDriver interface {
	net.Conn
	MakeRequest(node *Node, requestType string, data map[string]interface{}) *Request
	Request(request *Request)
	sendRequest(request *Request, retry int)
	Receive(receiveChannel chan Packet)
}