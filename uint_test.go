package stylus

import (
	"bytes"
	"encoding/hex"
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

func TestNewUint256(t *testing.T) {
	x := [32]byte(NewUint256(10112919))
	h, _ := hex.DecodeString("00000000000000000000000000000000000000000000000000000000009a4f97")
	if !bytes.Equal(x[:], h) {
		t.Fatalf("comp not correct: was %x", x)
	}
}
