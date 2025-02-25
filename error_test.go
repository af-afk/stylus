package stylus

import (
	"testing"
	"bytes"
	"encoding/hex"
)

func TestCoercion(t *testing.T) {
	b, ok := IsStylusErr(Err(1, "hello"))
	if !ok {
		t.Fatal("coercion not ok")
	}
	h, _ := hex.DecodeString("4e487b710000000000000000000000000000000000000000000000000000000000000001")
	if !bytes.Equal(b, h) {
		t.Fatalf("comp not correct: was: %x", b)
	}
}
