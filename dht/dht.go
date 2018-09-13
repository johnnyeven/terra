package dht

import (
	"net"
	"github.com/sirupsen/logrus"
	"github.com/ethereum/go-ethereum/p2p/nat"
	"math"
)

type DistributedHashTable struct {
	Network            string
	LocalAddr          string
	SeedNodes          []string
	conn               *Transport
	transactionManager *transactionManager
	nat                nat.Interface
	self               *Node
	packetChannel      chan Packet
	quitChannel        chan struct{}
	Handler            func(table *DistributedHashTable, packet Packet)
}

func (dht *DistributedHashTable) Run() {
	dht.init()
	dht.listen()
	dht.join()

	for {
		select {
		case packet := <-dht.packetChannel:
			dht.Handler(dht, packet)
		case <-dht.quitChannel:
			dht.conn.Close()
		}
	}
}

func (dht *DistributedHashTable) GetConn() *Transport {
	return dht.conn
}

func (dht *DistributedHashTable) GetTransactionManager() *transactionManager {
	return dht.transactionManager
}

func (dht *DistributedHashTable) init() {
	listener, err := net.ListenPacket(dht.Network, dht.LocalAddr)
	if err != nil {
		logrus.Panicf("[DistributedHashTable].init net.ListenPacket err: %v", err)
	}

	dht.conn = NewKRPCTransport(dht, listener.(*net.UDPConn))
	go dht.conn.Run()

	dht.transactionManager = newTransactionManager(math.MaxUint64)
	dht.nat = nat.Any()
	dht.packetChannel = make(chan Packet)
	dht.quitChannel = make(chan struct{})

	dht.self, err = NewNode(randomString(20), dht.Network, dht.LocalAddr)
	if err != nil {
		logrus.Panicf("[DistributedHashTable].init NewNode err: %v", err)
	}
	logrus.Info("initialized")
}

func (dht *DistributedHashTable) join() {
	for _, addr := range dht.SeedNodes {
		udpAddr, err := net.ResolveUDPAddr(dht.Network, addr)
		if err != nil {
			logrus.Warningf("[DistributedHashTable].join net.ResolveUDPAddr err: %v", err)
			continue
		}

		data := map[string]interface{}{
			"id":     dht.id(dht.self.id.RawString()),
			"target": dht.self.id.RawString(),
		}

		request := dht.conn.MakeRequest("find_node", data, udpAddr)
		dht.conn.Request(request)
		logrus.Info("send")
	}
}

func (dht *DistributedHashTable) listen() {
	realAddr := dht.conn.LocalAddr().(*net.UDPAddr)
	if dht.nat != nil {
		if !realAddr.IP.IsLoopback() {
			go nat.Map(dht.nat, dht.quitChannel, "udp4", realAddr.Port, realAddr.Port, "terra discovery")
		}
	}
	go dht.conn.Receive(dht.packetChannel)
	logrus.Info("listened")
}

func (dht *DistributedHashTable) id(target string) string {
	if target == "" {
		return dht.self.id.RawString()
	}
	return target[:15] + dht.self.id.RawString()[15:]
}
