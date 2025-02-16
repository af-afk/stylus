package stylus

import (
	"encoding/binary"
	"math/big"
)

var (
	// MaxUint256Big is 2^256 - 1.
	MaxUint256Big = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

	// MinInt256Big is -2^255.
	MinInt256Big = new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 255))

	// MaxInt256Big is 2^255 - 1.
	MaxInt256Big = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 255), big.NewInt(1))
)

// Uint256 word for internal accounting of storage accesses/etc.
type Uint256 [32]byte

// NewUint256 from a single word. Panics if it can't encode.
func NewUint256(u uint32) Uint256 {
	var x [32]byte
	if _, err := binary.Encode(x[32-4:], binary.BigEndian, u); err != nil {
		panic("encoding")
	}
	return Uint256(x)
}
