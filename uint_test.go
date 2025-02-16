package stylus

import (
	"math/big"
	"testing"
)

func TestMaxUint256BigShouldOverflow(t *testing.T) {
	var didPanic bool
	defer func() {
		recover()
		didPanic = true
	}()
	Uint256FromBig(new(big.Int).Add(
		MaxUint256Big,
		new(big.Int).SetInt64(100),
	))
	if !didPanic {
		t.Fail()
	}
}

func TestUint256(t *testing.T) {
	Uint256FromBig(new(big.Int).SetInt64(100))
}
