package bt

import (
	"github.com/johnnyeven/terra/dht"
	"github.com/sirupsen/logrus"
	"net"
	"fmt"
	"errors"
	"github.com/johnnyeven/terra/dht/util"
	"strings"
)

func BTHandlePacket(table *dht.DistributedHashTable, packet dht.Packet) {
	data, err := util.Decode(packet.Data)
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
		handler(table, packet.RemoteAddr.(*net.UDPAddr), response)
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
		errResponse := table.GetTransport().MakeError(nil, addr, tranID, dht.ProtocolError, err.Error())
		table.GetTransport().GetClient().(*dht.KRPCClient).Send(errResponse)
		return false
	}

	q := data["q"].(string)
	a := data["a"].(map[string]interface{})

	if err := dht.ParseKey(a, "id", "string"); err != nil {
		errResponse := table.GetTransport().MakeError(nil, addr, tranID, dht.ProtocolError, err.Error())
		table.GetTransport().GetClient().(*dht.KRPCClient).Send(errResponse)
		return false
	}

	id := a["id"].(string)
	if id == table.Self.ID.RawString() {
		return false
	}

	if len(id) != 20 {
		errResponse := table.GetTransport().MakeError(nil, addr, tranID, dht.ProtocolError, "invalid id length")
		table.GetTransport().GetClient().(*dht.KRPCClient).Send(errResponse)
		return false
	}

	if node, ok := table.GetRoutingTable().GetNodeByAddress(addr.String()); ok && node.ID.RawString() != id {
		table.GetRoutingTable().RemoveByAddr(addr.String())

		errResponse := table.GetTransport().MakeError(nil, addr, tranID, dht.ProtocolError, "invalid id")
		table.GetTransport().GetClient().(*dht.KRPCClient).Send(errResponse)
		return false
	}

	switch q {
	case dht.PingType:
		logrus.Info("ping request")
		response := table.GetTransport().MakeResponse(nil, addr, tranID, map[string]interface{}{"id": table.ID(id)})
		table.GetTransport().GetClient().(*dht.KRPCClient).Send(response)
		break
	case dht.FindNodeType:
		logrus.Info("find_node request")
		if err := dht.ParseKey(a, "target", "string"); err != nil {
			response := table.GetTransport().MakeError(nil, addr, tranID, dht.ProtocolError, err.Error())
			table.GetTransport().GetClient().(*dht.KRPCClient).Send(response)
			return false
		}

		target := a["target"].(string)
		if len(target) != 20 {
			response := table.GetTransport().MakeError(nil, addr, tranID, dht.ProtocolError, "invalid target")
			table.GetTransport().GetClient().(*dht.KRPCClient).Send(response)
			return false
		}

		var nodes string
		targetID := dht.NewIdentityFromString(target)

		no, _ := table.GetRoutingTable().GetNodeBucketByID(targetID)
		if no != nil {
			nodes = no.CompactNodeInfo()
		} else {
			nodes = strings.Join(
				table.GetRoutingTable().GetNeighborCompactInfos(targetID, table.K),
				"",
			)
		}

		data := map[string]interface{}{
			"id": table.ID(target),
			"nodes": nodes,
		}
		response := table.GetTransport().MakeResponse(nil, addr, tranID, data)
		table.GetTransport().GetClient().(*dht.KRPCClient).Send(response)

		break
	case dht.GetPeersType:
		logrus.Info("get_peers request")
		break
	case dht.AnnouncePeerType:
		logrus.Info("announce_peer request")
	}

	return true
}

func handleResponse(table *dht.DistributedHashTable, addr *net.UDPAddr, data map[string]interface{}) bool {
	tranID := data["t"].(string)
	tran := table.GetTransport().Get(tranID, addr)
	if tran == nil {
		return false
	}

	if err := dht.ParseKey(data, "r", "map"); err != nil {
		return false
	}

	q := tran.Data.(map[string]interface{})["q"].(string)
	a := tran.Data.(map[string]interface{})["a"].(map[string]interface{})
	r := data["r"].(map[string]interface{})

	if err := dht.ParseKey(r, "id", "string"); err != nil {
		return false
	}
	id := r["id"].(string)

	if tran.ClientID.(*dht.Identity) != nil && tran.ClientID.(*dht.Identity).RawString() != r["id"].(string) {
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

	if tran := table.GetTransport().Get(data["t"].(string), addr); tran != nil {
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
		logrus.Infof("new_node received, id: %x, ip: %s, port: %d", []byte(node.ID.RawString()), node.Addr.IP.String(), node.Addr.Port)
	}
	if found || !hasNew {
		return nil
	}

	id := targetID.RawData()
	for _, node := range table.GetRoutingTable().GetNeighbors(targetID, table.K) {
		switch requestType {
		case dht.FindNodeType:
			FindNode(node, table.GetTransport(), id)
		case dht.GetPeersType:
			GetPeers(node, table.GetTransport(), id)
		default:
			logrus.Panicf("[findOrContinueRequestTarget] err: invalid request type: %s", requestType)
		}
	}

	return nil
}
