package dht

import (
	"net"
	"errors"
)

type Node struct {
	id   *identity
	addr *net.UDPAddr
}

func NewNode(id, network, address string) (*Node, error) {
	if len(id) != 20 {
		return nil, errors.New("Node id should be a 20-length string")
	}

	addr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return nil, err
	}

	return &Node{newIdentityFromString(id), addr}, nil
}
