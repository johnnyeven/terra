package dht

import (
	"net"
	"errors"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	PingType         = "ping"
	FindNodeType     = "find_node"
	GetPeersType     = "get_peers"
	AnnouncePeerType = "announce_peer"
)

const (
	GeneralError = 201 + iota
	ServerError
	ProtocolError
	UnknownError
)

var _ interface {
	TransportDriver
} = (*KRPCClient)(nil)

type KRPCClient struct {
	conn *net.UDPConn
	dht  *DistributedHashTable
}

func (c *KRPCClient) MakeRequest(id *Identity, remoteAddr net.Addr, requestType string, data map[string]interface{}) *Request {
	params := MakeQuery(c.dht.transactionManager.generateTranID(), requestType, data)
	return &Request{
		ClientID:   id,
		cmd:        requestType,
		Data:       params,
		remoteAddr: remoteAddr,
	}
}

func (c *KRPCClient) MakeResponse(remoteAddr net.Addr, tranID string, data map[string]interface{}) *Request {
	params := MakeResponse(tranID, data)
	return &Request{
		Data:       params,
		remoteAddr: remoteAddr,
	}
}

func (c *KRPCClient) MakeError(remoteAddr net.Addr, tranID string, errCode int, errMsg string) *Request {
	params := MakeError(tranID, errCode, errMsg)
	return &Request{
		Data:       params,
		remoteAddr: remoteAddr,
	}
}

func (c *KRPCClient) Request(request *Request) {}

func (c *KRPCClient) sendRequest(request *Request, retry int) {
	tranID := request.Data["t"].(string)
	tran := c.dht.transactionManager.newTransaction(tranID, request, retry)
	c.dht.transactionManager.insertTransaction(tran)
	defer c.dht.transactionManager.deleteTransaction(tran.id)

	success := false
	for i := 0; i < retry; i++ {
		logrus.Warningf("[KRPCClient].Request c.conn.WriteToUDP try %d", i+1)
		err := c.Send(request)
		if err != nil {
			logrus.Warningf("[KRPCClient].Request c.conn.WriteToUDP err: %v", err)
			break
		}

		select {
		case <-tran.ResponseChannel:
			success = true
			break
		case <-time.After(time.Second * 15):
		}
	}

	if !success {
		c.dht.GetRoutingTable().RemoveByAddr(request.remoteAddr.String())
	}
}

func (c *KRPCClient) Send(request *Request) error {
	count, err := c.conn.WriteToUDP([]byte(Encode(request.Data)), request.remoteAddr.(*net.UDPAddr))
	if err != nil {
		return err
	}

	logrus.Debugf("[KRPCClient].Sent %d bytes", count)

	return nil
}

func (c *KRPCClient) Receive(receiveChannel chan Packet) {
	buff := make([]byte, 8192)
	for {
		n, raddr, err := c.conn.ReadFromUDP(buff)
		if err != nil {
			continue
		}

		receiveChannel <- Packet{buff[:n], raddr}
	}
}

func (c *KRPCClient) Read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}

func (c *KRPCClient) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

func (c *KRPCClient) Close() error {
	return c.conn.Close()
}

func (c *KRPCClient) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *KRPCClient) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *KRPCClient) SetDeadline(time time.Time) error {
	return c.conn.SetDeadline(time)
}

func (c *KRPCClient) SetReadDeadline(time time.Time) error {
	return c.conn.SetReadDeadline(time)
}

func (c *KRPCClient) SetWriteDeadline(time time.Time) error {
	return c.conn.SetWriteDeadline(time)
}

func ParseMessage(data interface{}) (map[string]interface{}, error) {
	response, ok := data.(map[string]interface{})
	if !ok {
		return nil, errors.New("response is not dict")
	}

	if err := ParseKeys(
		response, [][]string{{"t", "string"}, {"y", "string"}}); err != nil {
		return nil, err
	}

	return response, nil
}

func ParseKeys(data map[string]interface{}, pairs [][]string) error {
	for _, args := range pairs {
		key, t := args[0], args[1]
		if err := ParseKey(data, key, t); err != nil {
			return err
		}
	}
	return nil
}

func ParseKey(data map[string]interface{}, key string, t string) error {
	val, ok := data[key]
	if !ok {
		return errors.New("lack of key")
	}

	switch t {
	case "string":
		_, ok = val.(string)
	case "int":
		_, ok = val.(int)
	case "map":
		_, ok = val.(map[string]interface{})
	case "list":
		_, ok = val.([]interface{})
	default:
		panic("invalid type")
	}

	if !ok {
		return errors.New("invalid key type")
	}

	return nil
}

// makeQuery returns a query-formed data.
func MakeQuery(t, q string, a map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"t": t,
		"y": "q",
		"q": q,
		"a": a,
	}
}

// makeResponse returns a response-formed data.
func MakeResponse(t string, r map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"t": t,
		"y": "r",
		"r": r,
	}
}

// makeError returns a err-formed data.
func MakeError(t string, errCode int, errMsg string) map[string]interface{} {
	return map[string]interface{}{
		"t": t,
		"y": "e",
		"e": []interface{}{errCode, errMsg},
	}
}
