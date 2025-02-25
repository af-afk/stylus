package stylus

import (
	"fmt"
	"encoding/binary"
)

// StylusError is allowed to progagate as bad return data if something goes wrong.
type StylusError struct {
	Code int
	Explanation string
}

// Err creates a new StylusError for printing and revert reasons.
func Err(code int, explanation string) *StylusError {
	return &StylusError{code, explanation}
}

// Error stringifies this error, but it won't be called in the WASM
// environment.
func (s StylusError) Error() string {
	if e := s.Explanation; e != "" {
		return fmt.Sprintf("revert: %v", e)
	}
	return fmt.Sprintf("revert: code %v", s.Code)
}

// ContractBytes will be the returned data that the contract will return
// if something goes wrong of the ABI encoded form Error(uint256).
func (s StylusError) ContractBytes() []byte {
	var x [36]byte
	copy(x[:4], []byte{0x4e, 0x48, 0x7b, 0x71})
	binary.BigEndian.PutUint64(x[36-8:], uint64(s.Code))
	return x[:]
}

func IsStylusErr(x error) (d []byte, ok bool) {
	s, ok := x.(*StylusError)
	if !ok {
		return nil, ok
	}
	return s.ContractBytes(), true
}
