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
	fmt.Println(data)
	return true
}

func handleResponse(table *dht.DistributedHashTable, addr *net.UDPAddr, data map[string]interface{}) bool {
	tranID := data["t"].(string)
	tran := table.GetTransactionManager().GetByTranID(tranID)
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
		// remove
		return false
	}

	_, err := dht.NewNode(id, addr.Network(), addr.String())
	if err != nil {
		return false
	}

	switch q {
	case dht.PingType:
		break
	case dht.FindNodeType:
		//fmt.Println("find_node response")
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
	return true
}

func handleError(table *dht.DistributedHashTable, addr *net.UDPAddr, data map[string]interface{}) bool {
	fmt.Println(data)
	return true
}

func findOrContinueRequestTarget(table *dht.DistributedHashTable, targetID *dht.Identity, data map[string]interface{}, requestType string) error {
	if err := dht.ParseKey(data, "nodes", "string"); err != nil {
		return err
	}
	nodes := data["nodes"].(string)
	if len(nodes) % 26 != 0 {
		return errors.New("the length of nodes should can be divided by 26")
	}

	hasNew, found := false, false
	for i := 0; i < len(nodes)/26; i++ {
		node, _ := dht.NewNodeFromCompactInfo(string(nodes[i*26:(i+1)*26]), table.Network)
		if node.ID.RawString() == targetID.RawString() {
			found = true
		}

		hasNew = true
	}
	if found || !hasNew {
		return nil
	}

	return nil
}
