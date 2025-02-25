package stylus

import (
	"encoding/binary"
	"math/big"
	"bytes"
)

// zeroBig to use for comparison and ease of reading.
var zeroBig = new(big.Int).SetInt64(0)

var (
	// MaxUint256Big is 2^256 - 1.
	MaxUint256Big = new(big.Int).SetBytes(bytes.Repeat([]byte{0xff}, 32))

	// MaxInt256Big is 2^255 - 1.
	MaxInt256Big = new(big.Int).SetBytes([]byte{
		0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	})

	// MinInt256Big is -2^255.
	MinInt256Big = new(big.Int).Neg(MaxInt256Big)
)

// Uint256 word for internal accounting of storage accesses/etc.
type Uint256 [32]byte

// NewUint256 from a single word. Panics if it can't encode.
func NewUint256(u uint32) Uint256 {
	var x [32]byte
	binary.BigEndian.PutUint32(x[32-4:], u)
	return Uint256(x)
}
