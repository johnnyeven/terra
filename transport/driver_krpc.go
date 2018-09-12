package transport

import (
	"net"
	"errors"
	"github.com/sirupsen/logrus"
	"time"
)

type Packet struct {
	Data  []byte
	Raddr *net.UDPAddr
}

var _ interface {
	TransportDriver
} = (*krpcClient)(nil)

type krpcClient struct {
	conn *net.UDPConn
}

func (c *krpcClient) MakeRequest(requestType string, data map[string]interface{}, target net.Addr) *Request {
	params := MakeQuery("1", requestType, data)
	return &Request{
		cmd:        requestType,
		data:       params,
		remoteAddr: target,
	}
}

func (c *krpcClient) Request(request *Request) error {
	_, err := c.conn.WriteToUDP([]byte(Encode(request.data)), request.remoteAddr.(*net.UDPAddr))
	if err != nil {
		logrus.Warningf("[krpcClient].Request c.conn.WriteToUDP err: %v", err)
		return err
	}
	return nil
}

func (c *krpcClient) Receive(receiveChannel chan Packet) {
	buff := make([]byte, 8192)
	for {
		n, raddr, err := c.conn.ReadFromUDP(buff)
		if err != nil {
			continue
		}

		receiveChannel <- Packet{buff[:n], raddr}
	}
}

func (c *krpcClient) Read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}

func (c *krpcClient) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

func (c *krpcClient) Close() error {
	return c.conn.Close()
}

func (c *krpcClient) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *krpcClient) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *krpcClient) SetDeadline(time time.Time) error {
	return c.conn.SetDeadline(time)
}

func (c *krpcClient) SetReadDeadline(time time.Time) error {
	return c.conn.SetReadDeadline(time)
}

func (c *krpcClient) SetWriteDeadline(time time.Time) error {
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
