package spider

import "github.com/johnnyeven/terra/dht"

func Ping(node *dht.Node, t *dht.Transport) {
	data := map[string]interface{}{
		"id": t.GetDHT().ID(node.ID.RawString()),
	}

	request := t.MakeRequest(node.ID, node.Addr, dht.PingType, data)
	t.Request(request)
}

func FindNode(node *dht.Node, t *dht.Transport, target []byte) {
	if len(target) == 0 {
		target = t.GetDHT().Self.ID.RawData()
	}
	data := map[string]interface{}{
		"id":     t.GetDHT().ID(string(target)),
		"target": string(target),
	}

	request := t.MakeRequest(node.ID, node.Addr, dht.FindNodeType, data)
	t.Request(request)
}

func GetPeers(node *dht.Node, t *dht.Transport, infoHash []byte) {
	data := map[string]interface{}{
		"id":        t.GetDHT().ID(string(infoHash)),
		"info_hash": infoHash,
	}

	request := t.MakeRequest(node.ID, node.Addr, dht.GetPeersType, data)
	t.Request(request)
}

func AnnouncePeer(node *dht.Node, t *dht.Transport, infoHash string, impliedPort, port int, token string) {
	data := map[string]interface{}{
		"id":           t.GetDHT().ID(node.ID.RawString()),
		"info_hash":    infoHash,
		"implied_port": impliedPort,
		"port":         port,
		"token":        token,
	}

	request := t.MakeRequest(node.ID, node.Addr, dht.AnnouncePeerType, data)
	t.Request(request)
}
