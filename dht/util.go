package dht

import "crypto/rand"

func randomString(size int) string {
	buff := make([]byte, size)
	rand.Read(buff)
	return string(buff)
}
