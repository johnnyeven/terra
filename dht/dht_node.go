package dht

import (
	"net"
	"errors"
	"time"
	"strings"
	"github.com/johnnyeven/terra/dht/util"
)

type Node struct {
	ID             *Identity    `json:"id"`
	Addr           *net.UDPAddr `json:"addr"`
	LastActiveTime time.Time    `json:"lastActiveTime"`
}

func (node *Node) CompactNodeInfo() string {
	return strings.Join([]string{
		node.ID.RawString(), node.CompactIPPortInfo(),
	}, "")
}

func (node *Node) CompactIPPortInfo() string {
	info, _ := util.EncodeCompactIPPortInfo(node.Addr.IP, node.Addr.Port)
	return info
}

func NewNode(id, network, address string) (*Node, error) {
	if len(id) != 20 {
		return nil, errors.New("node ID should be a 20-length string")
	}

	addr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return nil, err
	}

	return &Node{NewIdentityFromString(id), addr, time.Now()}, nil
}

func NewNodeFromCompactInfo(compactNodeInfo string, network string) (*Node, error) {
	if len(compactNodeInfo) != 26 {
		return nil, errors.New("compactNodeInfo should be a 26-length string")
	}

	id := compactNodeInfo[:20]
	ip, port, err := util.DecodeCompactIPPortInfo(compactNodeInfo[20:])
	if err != nil {
		return nil, err
	}

	return NewNode(id, network, util.GenerateAddress(ip.String(), port))
}
