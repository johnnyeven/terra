package dht

import (
	"net"
	"errors"
)

type node struct {
	id   *identity
	addr *net.UDPAddr
}

func newNode(id, network, address string) (*node, error) {
	if len(id) != 20 {
		return nil, errors.New("node id should be a 20-length string")
	}

	addr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return nil, err
	}

	return &node{newIdentityFromString(id), addr}, nil
}
