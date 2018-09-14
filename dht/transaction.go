package dht

import (
	"sync"
	"strings"
	"net"
)

type transaction struct {
	*Request
	id              string
	ResponseChannel chan struct{}
}

type transactionManager struct {
	*sync.RWMutex
	transactions *syncedMap
	index        *syncedMap
	cursor       uint64
	maxCursor    uint64
	dht          *DistributedHashTable
}

func newTransactionManager(dht *DistributedHashTable, maxCursor uint64) *transactionManager {
	return &transactionManager{
		RWMutex:      &sync.RWMutex{},
		transactions: newSyncedMap(),
		index:        newSyncedMap(),
		maxCursor:    maxCursor,
		dht:          dht,
	}
}

func (c *transactionManager) generateTranID() string {
	c.Lock()
	defer c.Unlock()

	c.cursor = (c.cursor + 1) % c.maxCursor
	return string(int2bytes(c.cursor))
}

func (c *transactionManager) newTransaction(id string, request *Request, retry int) *transaction {
	return &transaction{
		Request:         request,
		id:              id,
		ResponseChannel: make(chan struct{}, retry+1),
	}
}

func (c *transactionManager) genIndexKey(queryType, address string) string {
	return strings.Join([]string{queryType, address}, ":")
}

func (c *transactionManager) genIndexKeyByTrans(tran *transaction) string {
	return c.genIndexKey(tran.Data["q"].(string), tran.remoteAddr.String())
}

func (c *transactionManager) insertTransaction(tran *transaction) {
	c.Lock()
	defer c.Unlock()

	c.transactions.Set(tran.id, tran)
	c.index.Set(c.genIndexKeyByTrans(tran), tran)
}

func (c *transactionManager) deleteTransaction(id string) {
	v, ok := c.transactions.Get(id)
	if !ok {
		return
	}

	c.Lock()
	defer c.Unlock()

	tran := v.(*transaction)
	c.transactions.Delete(tran.id)
	c.index.Delete(c.genIndexKeyByTrans(tran))
}

func (c *transactionManager) transactionLength() int {
	return c.transactions.Len()
}

func (c *transactionManager) transaction(key string, keyType int) *transaction {
	source := c.transactions
	if keyType == 1 {
		source = c.index
	}

	v, ok := source.Get(key)
	if !ok {
		return nil
	}

	return v.(*transaction)
}

func (c *transactionManager) GetByTranID(tranID string) *transaction {
	return c.transaction(tranID, 0)
}

func (c *transactionManager) GetByIndex(index string) *transaction {
	return c.transaction(index, 1)
}

func (c *transactionManager) Get(transID string, addr *net.UDPAddr) *transaction {
	trans := c.GetByTranID(transID)

	if trans == nil || trans.remoteAddr.String() != addr.String() {
		return nil
	}

	return trans
}

func (c *transactionManager) Ping(node *Node) {
	data := map[string]interface{}{
		"id": c.dht.ID(node.ID.RawString()),
	}

	request := c.dht.conn.MakeRequest(node.ID, node.addr, PingType, data)
	c.dht.conn.Request(request)
}

func (c *transactionManager) FindNode(node *Node, target string) {
	data := map[string]interface{}{
		"id":     c.dht.ID(target),
		"target": target,
	}

	request := c.dht.conn.MakeRequest(node.ID, node.addr, FindNodeType, data)
	c.dht.conn.Request(request)
}

func (c *transactionManager) GetPeers(node *Node, infoHash string) {
	data := map[string]interface{}{
		"id":        c.dht.ID(infoHash),
		"info_hash": infoHash,
	}

	request := c.dht.conn.MakeRequest(node.ID, node.addr, GetPeersType, data)
	c.dht.conn.Request(request)
}

func (c *transactionManager) AnnouncePeer(node *Node, infoHash string, impliedPort, port int, token string) {
	data := map[string]interface{}{
		"id":           c.dht.ID(node.ID.RawString()),
		"info_hash":    infoHash,
		"implied_port": impliedPort,
		"port":         port,
		"token":        token,
	}

	request := c.dht.conn.MakeRequest(node.ID, node.addr, AnnouncePeerType, data)
	c.dht.conn.Request(request)
}
