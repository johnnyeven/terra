package spider

import (
	"git.profzone.net/terra/dht"
	"github.com/sirupsen/logrus"
	"net"
	"fmt"
	"errors"
)

func BTHandlePacket(table *dht.DistributedHashTable, packet dht.Packet) {
	data, err := dht.Decode(packet.Data)
	if err != nil {
		logrus.Errorf("Decode err: %v", err)
	}

	response, err := dht.ParseMessage(data)
	if err != nil {
		logrus.Errorf("ParseMessage err: %v", err)
	}

	if err := dht.ParseKey(response, "y", "string"); err != nil {
		return
	}

	if handler, ok := handlers[response["y"].(string)]; ok {
		handler(table, packet.RemoteAddr, response)
	}
}

type dhtHandler func(*dht.DistributedHashTable, *net.UDPAddr, map[string]interface{}) bool

var handlers = map[string]dhtHandler{
	"q": handleRequest,
	"r": handleResponse,
	"e": handleError,
}

func handleRequest(table *dht.DistributedHashTable, addr *net.UDPAddr, data map[string]interface{}) bool {
	tranID := data["t"].(string)

	if err := dht.ParseKeys(data, [][]string{{"q", "string"}, {"a", "map"}}); err != nil {
		errResponse := table.GetConn().MakeError(addr, tranID, dht.ProtocolError, err.Error())
		table.GetConn().GetClient().(*dht.KRPCClient).Send(errResponse)
		return false
	}

	q := data["q"].(string)
	a := data["a"].(map[string]interface{})

	id := a["id"].(string)
	if id == table.Self.ID.RawString() {
		return false
	}

	if len(id) != 20 {
		errResponse := table.GetConn().MakeError(addr, tranID, dht.ProtocolError, "invalid id length")
		table.GetConn().GetClient().(*dht.KRPCClient).Send(errResponse)
		return false
	}

	switch q {
	case dht.PingType:
		logrus.Debug("ping request")
		response := table.GetConn().MakeResponse(addr, tranID, map[string]interface{}{"id": table.ID(id)})
		table.GetConn().GetClient().(*dht.KRPCClient).Send(response)
	case dht.FindNodeType:
	case dht.GetPeersType:
	case dht.AnnouncePeerType:

	}

	return true
}

func handleResponse(table *dht.DistributedHashTable, addr *net.UDPAddr, data map[string]interface{}) bool {
	tranID := data["t"].(string)
	tran := table.GetTransactionManager().Get(tranID, addr)
	if tran == nil {
		return false
	}

	if err := dht.ParseKey(data, "r", "map"); err != nil {
		return false
	}

	q := tran.Data["q"].(string)
	a := tran.Data["a"].(map[string]interface{})
	r := data["r"].(map[string]interface{})

	if err := dht.ParseKey(r, "id", "string"); err != nil {
		return false
	}
	id := r["id"].(string)

	if tran.ClientID != nil && tran.ClientID.RawString() != r["id"].(string) {
		table.GetRoutingTable().RemoveByAddr(addr.String())
		return false
	}

	node, err := dht.NewNode(id, addr.Network(), addr.String())
	if err != nil {
		return false
	}

	switch q {
	case dht.PingType:
		break
	case dht.FindNodeType:
		logrus.Debug("find_node response")
		target := a["target"].(string)
		if err := findOrContinueRequestTarget(table, dht.NewIdentityFromString(target), r, dht.FindNodeType); err != nil {
			return false
		}
	case dht.GetPeersType:
		fmt.Println("get_peers response")
	case dht.AnnouncePeerType:
		fmt.Println("ammounce_peer response")
	default:
		return false
	}

	tran.ResponseChannel <- struct{}{}
	table.GetRoutingTable().Insert(node)

	return true
}

func handleError(table *dht.DistributedHashTable, addr *net.UDPAddr, data map[string]interface{}) bool {
	if err := dht.ParseKey(data, "e", "list"); err != nil {
		return false
	}

	var e []interface{}
	if e = data["e"].([]interface{}); len(e) != 2 {
		return false
	}

	if tran := table.GetTransactionManager().Get(data["t"].(string), addr); tran != nil {
		tran.ResponseChannel <- struct{}{}
		logrus.Errorf("handled error errCode: %d, errMsg: %s", e[0].(int), e[1].(string))
	}
	return true
}

func findOrContinueRequestTarget(table *dht.DistributedHashTable, targetID *dht.Identity, data map[string]interface{}, requestType string) error {
	if err := dht.ParseKey(data, "nodes", "string"); err != nil {
		return err
	}
	nodes := data["nodes"].(string)
	if len(nodes)%26 != 0 {
		return errors.New("the length of nodes should can be divided by 26")
	}

	hasNew, found := false, false
	for i := 0; i < len(nodes)/26; i++ {
		node, _ := dht.NewNodeFromCompactInfo(string(nodes[i*26:(i+1)*26]), table.Network)
		if node.ID.RawString() == targetID.RawString() {
			found = true
		}

		if table.GetRoutingTable().Insert(node) {
			hasNew = true
		}
		logrus.Debugf("new_node received, %x", []byte(node.ID.RawString()))
	}
	if found || !hasNew {
		return nil
	}

	id := targetID.RawString()
	for _, node := range table.GetRoutingTable().GetNeighbors(targetID, table.K) {
		switch requestType {
		case dht.FindNodeType:
			table.GetTransactionManager().FindNode(node, id)
		case dht.GetPeersType:
		default:
			logrus.Panicf("[findOrContinueRequestTarget] err: invalid request type: %s", requestType)
		}
	}

	return nil
}
