package dht

import (
	"net"
	"github.com/sirupsen/logrus"
	"github.com/ethereum/go-ethereum/p2p/nat"
	"time"
	"git.profzone.net/terra/dht/util"
)

type DistributedHashTable struct {
	// the kbucket expired duration
	KBucketExpiredAfter time.Duration
	// the node expired duration
	NodeExpriedAfter time.Duration
	// how long it checks whether the bucket is expired
	CheckKBucketPeriod time.Duration
	// the max transaction id
	MaxTransactionCursor uint64
	// how many nodes routing table can hold
	MaxNodes int
	// in mainline dht, k = 8
	K int
	// for crawling mode, we put all nodes in one bucket, so KBucketSize may
	// not be K
	KBucketSize int
	// the nodes num to be fresh in a kbucket
	RefreshNodeNum int
	// udp, udp4, udp6
	Network       string
	// local network address
	LocalAddr     string
	// initialized node list
	SeedNodes     []string
	// the Transport communicating component
	transport     *Transport
	// node storage engin
	routingTable  *routingTable
	// NAT
	nat           nat.Interface
	// self node
	Self          *Node
	// received packet channel
	packetChannel chan Packet
	// system shutdown channel
	quitChannel   chan struct{}
	// packet handler
	Handler       func(table *DistributedHashTable, packet Packet)
}

func (dht *DistributedHashTable) Run() {
	dht.init()
	dht.listen()
	dht.join()

	tick := time.Tick(dht.CheckKBucketPeriod)

	for {
		select {
		case packet := <-dht.packetChannel:
			dht.Handler(dht, packet)
		case <-tick:
			if dht.routingTable.Len() == 0 {
				dht.join()
			} else if dht.transport.transactionLength() == 0 {
				go dht.routingTable.Fresh()
			}
		case <-dht.quitChannel:
			dht.transport.Close()
		}
	}
}

func (dht *DistributedHashTable) GetTransport() *Transport {
	return dht.transport
}

func (dht *DistributedHashTable) GetRoutingTable() *routingTable {
	return dht.routingTable
}

func (dht *DistributedHashTable) init() {
	listener, err := net.ListenPacket(dht.Network, dht.LocalAddr)
	if err != nil {
		logrus.Panicf("[DistributedHashTable].init net.ListenPacket err: %v", err)
	}

	dht.transport = NewKRPCTransport(dht, listener.(*net.UDPConn), dht.MaxTransactionCursor)
	go dht.transport.Run()

	dht.routingTable = newRoutingTable(dht.KBucketSize, dht)
	dht.nat = nat.Any()
	dht.packetChannel = make(chan Packet)
	dht.quitChannel = make(chan struct{})

	dht.Self, err = NewNode(util.RandomString(20), dht.Network, dht.LocalAddr)
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

		dht.transport.FindNode(&Node{addr: udpAddr}, dht.Self.ID.RawString())
	}
}

func (dht *DistributedHashTable) listen() {
	realAddr := dht.transport.LocalAddr().(*net.UDPAddr)
	if dht.nat != nil {
		if !realAddr.IP.IsLoopback() {
			go nat.Map(dht.nat, dht.quitChannel, "udp4", realAddr.Port, realAddr.Port, "terra discovery")
		}
	}
	go dht.transport.Receive(dht.packetChannel)
	logrus.Info("listened")
}

func (dht *DistributedHashTable) ID(target string) string {
	if target == "" {
		return dht.Self.ID.RawString()
	}
	return target[:15] + dht.Self.ID.RawString()[15:]
}
