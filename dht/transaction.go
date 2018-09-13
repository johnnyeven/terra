package dht

import "sync"

type transaction struct {
	*Request
	id              string
	ResponseChannel chan struct{}
}

type transactionManager struct {
	*sync.RWMutex
	transactions map[string]*transaction
	cursor       uint64
	maxCursor    uint64
	dht          *DistributedHashTable
}

func newTransactionManager(dht *DistributedHashTable, maxCursor uint64) *transactionManager {
	return &transactionManager{
		RWMutex:      &sync.RWMutex{},
		transactions: make(map[string]*transaction),
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

func (c *transactionManager) insertTransaction(tran *transaction) {
	c.Lock()
	defer c.Unlock()

	c.transactions[tran.id] = tran
}

func (c *transactionManager) deleteTransaction(id string) {
	if _, ok := c.transactions[id]; !ok {
		return
	}

	c.Lock()
	defer c.Unlock()

	delete(c.transactions, id)
}

func (c *transactionManager) transactionLength() int {
	return len(c.transactions)
}

func (c *transactionManager) GetByTranID(tranID string) *transaction {
	return c.transactions[tranID]
}

func (c *transactionManager) FindNode(no *Node, target string) {
	data := map[string]interface{}{
		"id":     c.dht.id(target),
		"target": target,
	}

	request := c.dht.conn.MakeRequest(no, "find_node", data)
	c.dht.conn.Request(request)
}
