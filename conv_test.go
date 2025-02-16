package stylus

import (
	"math/big"
	"encoding/hex"
	"testing"
)

func TestBigToInt256Bytes(t *testing.T) {
	x := BigToInt256Bytes(new(big.Int).SetInt64(-100))
	if s := hex.EncodeToString(x); s != "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff9c" {
		t.Logf("neg -100: %s", s)
		t.Fail()
	}
	ib, _ := hex.DecodeString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff9c")
	i, err := BInt256ToBig(ib)
	if err != nil {
		t.Fatalf("decode str: %v", err)
	}
	if i.Cmp(new(big.Int).SetInt64(-100)) != 0 {
		t.Fail()
	}
	x = BigToInt256Bytes(new(big.Int).SetInt64(102811))
	if s := hex.EncodeToString(x); s != "000000000000000000000000000000000000000000000000000000000001919b" {
		t.Logf("102811: %s", s)
		t.Fail()
	}
}
