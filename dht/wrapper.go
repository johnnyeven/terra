package dht

import (
	"net"
	"time"
)

type Transport struct {
	client         TransportDriver
	requestChannel chan *Request
}

var _ interface {
	TransportDriver
} = (*Transport)(nil)

func (t *Transport) GetClient() TransportDriver {
	return t.client
}

func (t *Transport) Run() {
	for {
		select {
		case r := <-t.requestChannel:
			go t.sendRequest(r, 5)
		}
	}
}

func (t *Transport) sendRequest(request *Request, retry int) {
	t.client.sendRequest(request, retry)
}

func (t *Transport) Init(client TransportDriver) {
	t.client = client
	t.requestChannel = make(chan *Request)
}

func (t *Transport) MakeRequest(id *Identity, remoteAddr net.Addr, requestType string, data map[string]interface{}) *Request {
	return t.client.MakeRequest(id, remoteAddr, requestType, data)
}

func (t *Transport) MakeResponse(remoteAddr net.Addr, tranID string, data map[string]interface{}) *Request {
	return t.client.MakeResponse(remoteAddr, tranID, data)
}

func (t *Transport) MakeError(remoteAddr net.Addr, tranID string, errCode int, errMsg string) *Request {
	return t.client.MakeError(remoteAddr, tranID, errCode, errMsg)
}

func (t *Transport) Request(request *Request) {
	t.requestChannel <- request
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

func NewKRPCTransport(dht *DistributedHashTable, conn *net.UDPConn) *Transport {
	trans := &Transport{}
	trans.Init(&KRPCClient{
		dht:  dht,
		conn: conn,
	})

	return trans
}
