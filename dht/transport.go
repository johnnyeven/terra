package dht

import (
	"net"
	"time"
	"sync"
	"strings"
	"git.profzone.net/profzone/terra/dht/util"
)

const RequestRetryTime = 2

type transaction struct {
	*Request
	id              string
	ResponseChannel chan struct{}
}

type Transport struct {
	TransportDriver
	*sync.RWMutex
	transactions   *SyncedMap
	index          *SyncedMap
	cursor         uint64
	maxCursor      uint64
	dht            *DistributedHashTable
	client         TransportDriver
	requestChannel chan *Request
	quitChannel    chan struct{}
}

var _ interface {
	TransportDriver
} = (*Transport)(nil)

func (t *Transport) GetDHT() *DistributedHashTable {
	return t.dht
}

func (t *Transport) generateTranID() string {
	t.Lock()
	defer t.Unlock()

	t.cursor = (t.cursor + 1) % t.maxCursor
	return string(util.Int2Bytes(t.cursor))
}

func (t *Transport) newTransaction(id string, request *Request, retry int) *transaction {
	return &transaction{
		Request:         request,
		id:              id,
		ResponseChannel: make(chan struct{}, retry+1),
	}
}

func (t *Transport) genIndexKey(queryType, address string) string {
	return strings.Join([]string{queryType, address}, ":")
}

func (t *Transport) genIndexKeyByTrans(tran *transaction) string {
	return t.genIndexKey(tran.Data.(map[string]interface{})["q"].(string), tran.remoteAddr.String())
}

func (t *Transport) insertTransaction(tran *transaction) {
	t.Lock()
	defer t.Unlock()

	t.transactions.Set(tran.id, tran)
	t.index.Set(t.genIndexKeyByTrans(tran), tran)
}

func (t *Transport) deleteTransaction(id string) {
	v, ok := t.transactions.Get(id)
	if !ok {
		return
	}

	t.Lock()
	defer t.Unlock()

	tran := v.(*transaction)
	t.transactions.Delete(tran.id)
	t.index.Delete(t.genIndexKeyByTrans(tran))
}

func (t *Transport) transactionLength() int {
	return t.transactions.Len()
}

func (t *Transport) transaction(key string, keyType int) *transaction {
	source := t.transactions
	if keyType == 1 {
		source = t.index
	}

	v, ok := source.Get(key)
	if !ok {
		return nil
	}

	return v.(*transaction)
}

func (t *Transport) GetByTranID(tranID string) *transaction {
	return t.transaction(tranID, 0)
}

func (t *Transport) GetByIndex(index string) *transaction {
	return t.transaction(index, 1)
}

func (t *Transport) Get(transID string, addr *net.UDPAddr) *transaction {
	trans := t.GetByTranID(transID)

	if trans == nil || trans.remoteAddr.String() != addr.String() {
		return nil
	}

	return trans
}

func (t *Transport) GetClient() TransportDriver {
	return t.client
}

func (t *Transport) Run() {

Run:
	for {
		select {
		case r := <-t.requestChannel:
			go t.SendRequest(r, RequestRetryTime)
		case <-t.quitChannel:
			break Run
		}
	}
}

func (t *Transport) SendRequest(request *Request, retry int) {
	t.client.SendRequest(request, retry)
}

func (t *Transport) Init(table *DistributedHashTable, client TransportDriver, maxCursor uint64) {
	t.client = client
	t.requestChannel = make(chan *Request)
	t.quitChannel = make(chan struct{})
	t.RWMutex = &sync.RWMutex{}
	t.transactions = NewSyncedMap()
	t.index = NewSyncedMap()
	t.maxCursor = maxCursor
	t.dht = table
}

func (t *Transport) MakeRequest(id interface{}, remoteAddr net.Addr, requestType string, data interface{}) *Request {
	return t.client.MakeRequest(id, remoteAddr, requestType, data)
}

func (t *Transport) MakeResponse(remoteAddr net.Addr, tranID string, data interface{}) *Request {
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
	t.quitChannel <- struct{}{}
	close(t.requestChannel)
	close(t.quitChannel)
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
