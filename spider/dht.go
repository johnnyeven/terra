package spider

import (
	"net"
	"github.com/sirupsen/logrus"
	"git.profzone.net/terra/transport/krpc"
	"fmt"
)

type DistributedHashTable struct {
	Network       string
	LocalAddr     string
	SeedNodes     []string
	conn          *net.UDPConn
	self          *node
	packetChannel chan krpc.Packet
}

func (dht *DistributedHashTable) Run() {
	dht.init()
	dht.join()

	for {
		select {
		case packet := <- dht.packetChannel:
			data, err := Decode(packet.Data)
			if err != nil {
				logrus.Errorf("Decode err: %v", err)
			}

			response, err := krpc.ParseMessage(data)
			if err != nil {
				logrus.Errorf("ParseMessage err: %v", err)
			}

			fmt.Println(response)
		}
	}
}

func (dht *DistributedHashTable) init() {
	listener, err := net.ListenPacket(dht.Network, dht.LocalAddr)
	if err != nil {
		logrus.Panicf("[DistributedHashTable].init net.ListenPacket err: %v", err)
	}

	dht.conn = listener.(*net.UDPConn)
	dht.packetChannel = make(chan krpc.Packet)

	dht.self, err = newNode(randomString(20), dht.Network, dht.LocalAddr)
	if err != nil {
		logrus.Panicf("[DistributedHashTable].init newNode err: %v", err)
	}
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

		_, err = dht.conn.WriteToUDP([]byte(Encode(data)), udpAddr)
		if err != nil {
			logrus.Warningf("[DistributedHashTable].join dht.conn.WriteToUDP err: %v", err)
		}
	}
}

func (dht *DistributedHashTable) listen() {
	go func() {
		buff := make([]byte, 8192)
		for {
			n, raddr, err := dht.conn.ReadFromUDP(buff)
			if err != nil {
				continue
			}

			dht.packetChannel <- krpc.Packet{buff[:n], raddr}
		}
	}()
}

func (dht *DistributedHashTable) id(target string) string {
	if target == "" {
		return dht.self.id.RawString()
	}
	return target[:15] + dht.self.id.RawString()[15:]
}
