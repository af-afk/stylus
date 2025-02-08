package main

//go:generate stylus-go

import (
	"math/big"

	"github.com/af-afk/stylus"
)

//stylus entrypoint
type Storage struct {
	Counter stylus.StorageUint256
}

//stylus uint256
func (s Storage) Burn(x *big.Int) ([]byte, error) {
	y := new(big.Int).SetInt64(123)
	y.Add(x, y)
	var b [32]byte
	x.FillBytes(b[:])
	return b[:], nil
}
