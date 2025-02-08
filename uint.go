package stylus

import (
	"encoding/binary"
	"math/big"
)

var maxUint32Word = big.Word(^uint32(0))

// MaxUint256Big that could be created.
var MaxUint256Big = new(big.Int).SetBits([]big.Word{
	// We need 32 bit words here, since that's all we can use with
	// the wasm backend.
	maxUint32Word,
	maxUint32Word,
	maxUint32Word,
	maxUint32Word,
	maxUint32Word,
	maxUint32Word,
	maxUint32Word,
	maxUint32Word,
})

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

// Uint256FromBig for internal storage map use presumably. Panics if
// overflowing (and underflowing) so use with caution.
func Uint256FromBig(x *big.Int) (u Uint256) {
	if x.Cmp(new(big.Int)) < 0 {
		panic("underflow")
	}
	if x.Cmp(MaxUint256Big) > 0 {
		panic("overflow")
	}
	var b [32]byte
	return Uint256(x.FillBytes(b[:]))
}

// BigFromUint256 type conversion.
func BigFromUint256(x Uint256) *big.Int {
	return new(big.Int).SetBytes(x[:])
}
