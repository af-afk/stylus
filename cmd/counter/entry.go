package main

//go:generate stylus-go

import (
	"math/big"
	"fmt"

	"github.com/af-afk/stylus"
)

//stylus entrypoint
type Storage struct {
	Counter stylus.StorageUint256
}

//stylus uint256
func (s Storage) Add(x *big.Int) ([]byte, error) {
	y := s.Counter.Get()
	y.Add(y, x)
	s.Counter.Set(y)
	var b [32]byte
	x.FillBytes(b[:])
	return b[:], nil
}

func (s Storage) Sub(x *big.Int) ([]byte, error) {
	y := s.Counter.Get()
	y.Sub(y, x)
	s.Counter.Set(y)
	var b [32]byte
	x.FillBytes(b[:])
	return b[:], nil
}

func (s Storage) Count() (b []byte, err error) {
	b = make([]byte, 32)
	s.Counter.Get().FillBytes(b)
	return
}

func (s Storage) FakeDie() (b []byte, err error) {
	msg := "uhoh!"
	return []byte(msg), fmt.Errorf(msg)
}
