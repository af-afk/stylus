package stylus

import "math/big"

func BInt256ToBig(x []byte) (*big.Int, error) {
	// Check the msb to see if we're a negative number.
	if x[0]&0x80 == 0 {
		// This is a positive number.
		return new(big.Int).SetBytes(x), nil
	}
	t := new(big.Int).SetBytes(x)
	i := t.Sub(t, new(big.Int).Lsh(big.NewInt(1), 256))
	return i, nil
}

func BUint256ToBig(x []byte) (*big.Int, error) {
	return new(big.Int).SetBytes(x), nil
}

// BigToUint256Bytes by filling out a 32 byte word, panicking if the word
// would overflow.
func BigToUint256Bytes(x *big.Int) [32]byte {
	// Check if an underflow or overflow has taken place.
	if x.Cmp(zeroBig) < 0 || x.Cmp(MaxUint256Big) > 0 {
		setRdUnderOverflow()
		panic("under/overflow")
	}
	var b [32]byte
	x.FillBytes(b[:])
	return b
}

// BigToInt256Bytes by filling out a 32 byte word, panicking if the word
// would overflow.
func BigToInt256Bytes(i *big.Int) []byte {
	// Check if underflowing/overflowing.
	if i.Cmp(MinInt256Big) < 0 || i.Cmp(MaxInt256Big) > 0 {
		setRdUnderOverflow()
		panic("underflow/overflow")
	}
	var b [32]byte
	if i.Sign() < 0 {
		// Do some work for twos complement here.
		t := new(big.Int).Add(i, new(big.Int).Lsh(big.NewInt(1), 256))
		x := t.Bytes()
		copy(b[32-len(x):], x)
		for j := 0; j < 32-len(x); j++ {
			b[j] = 0xff
		}
	} else {
		i.FillBytes(b[:])
	}
	return b[:]
}

func Uint256ToBytes(x Uint256) []byte {
	return x[:]
}

// Uint256FromBig for internal storage map use presumably. Panics if
// overflowing (and underflowing) so use with caution.
func Uint256FromBig(x *big.Int) (u Uint256) {
	return Uint256(BigToUint256Bytes(x))
}

// BigFromUint256 type conversion.
func BigFromUint256(x Uint256) *big.Int {
	return new(big.Int).SetBytes(x[:])
}

// U is an abbreviated form of Uint256FromBig for simplicity's sake.
func U(x *big.Int) Uint256 {
	return Uint256FromBig(x)
}

// BytesIdentity is a helper function for return type matching to return
// the slice as itself.
func BytesIdentity(b []byte) []byte {
	return b
}
