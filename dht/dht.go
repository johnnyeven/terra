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
	Self               *Node
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

	dht.transactionManager = newTransactionManager(dht, math.MaxUint64)
	dht.nat = nat.Any()
	dht.packetChannel = make(chan Packet)
	dht.quitChannel = make(chan struct{})

	dht.Self, err = NewNode(randomString(20), dht.Network, dht.LocalAddr)
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

		dht.transactionManager.FindNode(&Node{addr: udpAddr}, dht.Self.ID.RawString())
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

func (dht *DistributedHashTable) ID(target string) string {
	if target == "" {
		return dht.Self.ID.RawString()
	}
	return target[:15] + dht.Self.ID.RawString()[15:]
}
