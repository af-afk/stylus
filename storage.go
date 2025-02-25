package stylus

import (
	"math/big"

	"github.com/af-afk/stylus/host"
)

// storageCounter, useful for an internal base count storage locations.
var storageCounter uint32 = 0

type (
	// StorageUint256 containing a uint256, loaded eagerly.
	StorageUint256 struct{ o Uint256 }

	// StorageInt256 containing a int256, loaded eagerly.
	StorageInt256 struct{ o Uint256 }

	// StorageAddress containing an address, loaded eagerly.
	StorageAddress struct { o Uint256 }
)

// NewStorageUint256 returns a free storage pointer for uint256.
func NewStorageUint256() StorageUint256 {
	x := StorageUint256{NewUint256(storageCounter)}
	storageCounter++
	return x
}

func (s StorageUint256) Get() *big.Int {
	return LoadStorageBig(&s.o)
}

func (s StorageUint256) Set(x *big.Int) {
	b := Uint256FromBig(x)
	SetStorageUint256(&s.o, &b)
}

// LoadStorageWord using the offset given.
func LoadStorageWord(offset *Uint256) (x [32]byte) {
	var w [32]byte
	b := &w
	host.StorageLoadBytes32(unsafePtr(&offset), unsafePtr(&b))
	return w
}

// LoadStorageBig from memory, using a fixed-size storage for it.
func LoadStorageBig(offset *Uint256) (i *big.Int) {
	i = BigFromUint256(LoadStorageWord(offset))
	return
}

func SetStorageUint256(offset *Uint256, v *Uint256) {
	host.StorageCacheBytes32(unsafePtr(&offset), unsafePtr(&v))
}

// NewStorageInt256 returns a free storage pointer for int256.
func NewStorageInt256() StorageInt256 {
	x := StorageInt256{NewUint256(storageCounter)}
	storageCounter++
	return x
}

func (s StorageInt256) Get() *big.Int {
	w := LoadStorageWord(&s.o)
	x, err :=  BytesToInt256Big(w[:])
	if err != nil {
		setRdUnderOverflow()
		panic("bad word")
	}
	return x
}

func (s StorageInt256) Set(x *big.Int) {
	b := BigToInt256Bytes(x)
	host.StorageCacheBytes32(unsafePtr(&s.o), unsafePtr(&b))
}

func NewStorageAddress() StorageAddress {
	x := StorageAddress{NewUint256(storageCounter)}
	storageCounter++
	return x
}

func (s StorageAddress) Get() Address {
	x := LoadStorageWord(&s.o)
	a, err := BytesToAddress(x[12:])
	if err != nil {
		panic("bad addr")
	}
	return a
}

func (s StorageAddress) Set(x Address) {
	a := AddressToWord(x)
	host.StorageCacheBytes32(unsafePtr(&s.o), unsafePtr(&a))
}

