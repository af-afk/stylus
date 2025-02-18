package main

//go:generate stylus-go gen

import (
	"math/big"

	"github.com/af-afk/stylus"
)

//stylus entrypoint
type Storage struct {
	Counter stylus.StorageUint256
}

//stylus uint256
func (s Storage) Add(x *big.Int) (*big.Int, error) {
	y := s.Counter.Get()
	y.Add(y, x)
	s.Counter.Set(y)
	return y, nil
}

func (s Storage) Sub(x *big.Int) (stylus.Uint256, error) {
	y := s.Counter.Get()
	y.Sub(y, x)
	s.Counter.Set(y)
	return stylus.Uint256FromBig(y), nil
}

func (s Storage) Count() (b *big.Int, err error) {
	return s.Counter.Get(), nil
}
