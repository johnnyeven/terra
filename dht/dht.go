package dht

import (
	"net"
	"github.com/sirupsen/logrus"
	"github.com/ethereum/go-ethereum/p2p/nat"
	"git.profzone.net/terra/transport"
)

type DistributedHashTable struct {
	Network       string
	LocalAddr     string
	SeedNodes     []string
	conn          *transport.Transport
	nat           nat.Interface
	self          *node
	packetChannel chan transport.Packet
	quitChannel   chan struct{}
	Handler       func(table *DistributedHashTable, packet transport.Packet)
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

func (dht *DistributedHashTable) init() {
	listener, err := net.ListenPacket(dht.Network, dht.LocalAddr)
	if err != nil {
		logrus.Panicf("[DistributedHashTable].init net.ListenPacket err: %v", err)
	}

	dht.conn = transport.NewKRPCTransport(listener.(*net.UDPConn))
	dht.nat = nat.Any()
	dht.packetChannel = make(chan transport.Packet)
	dht.quitChannel = make(chan struct{})

	dht.self, err = newNode(randomString(20), dht.Network, dht.LocalAddr)
	if err != nil {
		logrus.Panicf("[DistributedHashTable].init newNode err: %v", err)
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
		err = dht.conn.Request(request)
		if err != nil {
			logrus.Warningf("[DistributedHashTable].join dht.conn.WriteToUDP err: %v", err)
		}
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
