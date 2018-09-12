package dht

import (
	"github.com/sirupsen/logrus"
	"fmt"
	"strings"
)

type identity struct {
	Size int
	data []byte
}

func (id *identity) Bit(index int) int {
	if index >= id.Size {
		logrus.Panic("[identity].Bit err: index out of range")
	}

	div, mod := index/8, index%8
	return int((uint(id.data[div]) & (1 << uint(7-mod))) >> uint(7-mod))
}

func (id *identity) set(index int, value int) {
	if index >= id.Size {
		logrus.Panic("[identity].set err: index out of range")
	}

	div, mod := index/8, index%8
	shift := byte(1 << uint(7-mod))

	id.data[div] &= ^shift
	if value > 0 {
		id.data[div] |= shift
	}
}

func (id *identity) Set(index int) {
	id.set(index, 1)
}

func (id *identity) Unset(index int) {
	id.set(index, 0)
}

func (id *identity) Compare(source *identity, prefixLen int) int {
	if prefixLen > id.Size || prefixLen > source.Size {
		logrus.Panic("[identity].Compare err: index out of range")
	}

	div, mod := prefixLen/8, prefixLen%8
	for i := 0; i < div; i++ {
		if id.data[i] > source.data[i] {
			return 1
		} else if id.data[i] < source.data[i] {
			return -1
		}
	}

	for i := div * 8; i < div*8+mod; i++ {
		bit1, bit2 := id.Bit(i), source.Bit(i)
		if bit1 > bit2 {
			return 1
		} else if bit1 < bit2 {
			return -1
		}
	}

	return 0
}

func (id *identity) Xor(source *identity) *identity {
	if id.Size != source.Size {
		logrus.Panic("[identity].Xor err: size not the same")
	}

	distance := newIdentity(id.Size)
	div, mod := distance.Size/8, distance.Size%8

	for i := 0; i < div; i++ {
		distance.data[i] = id.data[i] ^ source.data[i]
	}

	for i := div * 8; i < div*8+mod; i++ {
		distance.set(i, id.Bit(i)^source.Bit(i))
	}

	return distance
}

func (id *identity) String() string {
	div, mod := id.Size/8, id.Size%8
	buff := make([]string, div+mod)

	for i := 0; i < div; i++ {
		buff[i] = fmt.Sprintf("%08b", id.data[i])
	}

	for i := div; i < div+mod; i++ {
		buff[i] = fmt.Sprintf("%1b", id.Bit(div*8+(i-div)))
	}

	return strings.Join(buff, "")
}

func (id *identity) RawString() string {
	return string(id.data)
}

func newIdentity(size int) *identity {
	div := size / 8
	mod := size % 8

	if mod > 0 {
		div++
	}

	return &identity{
		Size: size,
		data: make([]byte, div),
	}
}

func newIdentityCopy(source *identity, size int) *identity {
	target := newIdentity(size)

	if size > source.Size {
		size = source.Size
	}

	div := size / 8

	for i := 0; i < div; i++ {
		target.data[i] = source.data[i]
	}

	for i := div * 8; i < size; i++ {
		if source.Bit(i) == 1 {
			target.Set(i)
		}
	}

	return target
}

func newIdentityFromBytes(data []byte) *identity {
	target := newIdentity(len(data) * 8)
	copy(target.data, data)

	return target
}

func newIdentityFromString(data string) *identity {
	return newIdentityFromBytes([]byte(data))
}
