package dht

import (
	"crypto/rand"
	"net"
	"errors"
	"strings"
	"strconv"
)

func randomString(size int) string {
	buff := make([]byte, size)
	rand.Read(buff)
	return string(buff)
}

func int2bytes(val uint64) []byte {
	data, j := make([]byte, 8), -1
	for i := 0; i < 8; i++ {
		shift := uint64((7 - i) * 8)
		data[i] = byte((val & (0xff << shift)) >> shift)

		if j == -1 && data[i] != 0 {
			j = i
		}
	}

	if j != -1 {
		return data[j:]
	}
	return data[:1]
}

func decodeCompactIPPortInfo(info string) (ip net.IP, port int, err error) {
	if len(info) != 6 {
		err = errors.New("compact info should be 6-length long")
		return
	}

	ip = net.IPv4(info[0], info[1], info[2], info[3])
	port = int((uint16(info[4]) << 8) | uint16(info[5]))
	return
}

func encodeCompactIPPortInfo(ip net.IP, port int) (info string, err error) {
	if port > 65535 || port < 0 {
		err = errors.New(
			"port should be no greater than 65535 and no less than 0")
		return
	}

	p := int2bytes(uint64(port))
	if len(p) < 2 {
		p = append(p, p[0])
		p[0] = 0
	}

	info = string(append(ip, p...))
	return
}

func genAddress(ip string, port int) string {
	return strings.Join([]string{ip, strconv.Itoa(port)}, ":")
}

