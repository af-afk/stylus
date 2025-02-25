package stylus

import (
	"bytes"
	"encoding/hex"
	"strings"
)

type Address [20]byte

// AddressFromString, useful for hardcoded addresses. Will panic if it
// can't unpack the address at runtime.
func AddressFromString(s string) Address {
	b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil {
		panic(err)
	}
	a, err := BytesToAddress(b)
	if err != nil {
		panic(err)
	}
	return a
}

func (x Address) Equal(y Address) bool {
	return bytes.Equal(x[:], y[:])
}

func (a Address) Bytes() []byte {
	return a[:]
}
