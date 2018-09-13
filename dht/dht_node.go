package dht

import (
	"net"
	"errors"
)

type Node struct {
	ID   *Identity
	addr *net.UDPAddr
}

func NewNode(id, network, address string) (*Node, error) {
	if len(id) != 20 {
		return nil, errors.New("node ID should be a 20-length string")
	}

	addr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return nil, err
	}

	return &Node{NewIdentityFromString(id), addr}, nil
}

func NewNodeFromCompactInfo(compactNodeInfo string, network string) (*Node, error) {
	if len(compactNodeInfo) != 26 {
		return nil, errors.New("compactNodeInfo should be a 26-length string")
	}

	id := compactNodeInfo[:20]
	ip, port, err := decodeCompactIPPortInfo(compactNodeInfo[20:])
	if err != nil {
		return nil, err
	}

	return NewNode(id, network, genAddress(ip.String(), port))
}
