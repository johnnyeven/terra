package spider

import (
	"git.profzone.net/terra/transport"
	"github.com/sirupsen/logrus"
	"git.profzone.net/terra/dht"
	"net"
)

func BTHandlePacket(dht *dht.DistributedHashTable, packet transport.Packet) {
	data, err := transport.Decode(packet.Data)
	if err != nil {
		logrus.Errorf("Decode err: %v", err)
	}

	response, err := transport.ParseMessage(data)
	if err != nil {
		logrus.Errorf("ParseMessage err: %v", err)
	}

	if handler, ok := handlers[response["y"].(string)]; ok {
		handler(dht, packet.Raddr, response)
	}
}

type dhtHandler func(*dht.DistributedHashTable, *net.UDPAddr, map[string]interface{}) bool

var handlers = map[string]dhtHandler{
	"q": handleRequest,
	"r": handleResponse,
	"e": handleError,
}

func handleRequest(table *dht.DistributedHashTable, addr *net.UDPAddr, data map[string]interface{}) bool {
	return true
}

func handleResponse(table *dht.DistributedHashTable, addr *net.UDPAddr, data map[string]interface{}) bool {
	return true
}

func handleError(table *dht.DistributedHashTable, addr *net.UDPAddr, data map[string]interface{}) bool {
	return true
}
