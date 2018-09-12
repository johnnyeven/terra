package transport

import (
	"net"
	"time"
)

type Transport struct {
	client TransportDriver
}

var _ interface {
	TransportDriver
} = (*Transport)(nil)

func (t *Transport) Init(client TransportDriver) {
	t.client = client
}

func (t *Transport) MakeRequest(requestType string, data map[string]interface{}, target net.Addr) *Request {
	return t.client.MakeRequest(requestType, data, target)
}

func (t *Transport) Request(request *Request) error {
	return t.client.Request(request)
}

func (t *Transport) Receive(receiveChannel chan Packet) {
	t.client.Receive(receiveChannel)
}

func (t *Transport) Read(b []byte) (n int, err error) {
	return t.client.Read(b)
}

func (t *Transport) Write(b []byte) (n int, err error) {
	return t.client.Write(b)
}

func (t *Transport) Close() error {
	return t.client.Close()
}

func (t *Transport) LocalAddr() net.Addr {
	return t.client.LocalAddr()
}

func (t *Transport) RemoteAddr() net.Addr {
	return t.client.RemoteAddr()
}

func (t *Transport) SetDeadline(time time.Time) error {
	return t.client.SetDeadline(time)
}

func (t *Transport) SetReadDeadline(time time.Time) error {
	return t.client.SetReadDeadline(time)
}

func (t *Transport) SetWriteDeadline(time time.Time) error {
	return t.client.SetWriteDeadline(time)
}

func NewKRPCTransport(conn *net.UDPConn) *Transport {
	trans := &Transport{}
	trans.Init(&krpcClient{conn})

	return trans
}
