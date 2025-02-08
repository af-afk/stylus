package stylus

import "math/big"

// storageCounter, useful for an internal count of the storage locations.
var storageCounter uint32 = 0

// StorageUint256 slot, containing its offset.
type StorageUint256 struct{ o Uint256 }

// NewStorageUint256 returns a free storage pointer to be used in this program.
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

// LoadStorageBig from memory, using a fixed-size storage for it.
func LoadStorageBig(offset *Uint256) (i *big.Int) {
	var w [32]byte
	b := &w
	StorageLoadBytes32(unsafePtr(&offset), unsafePtr(&b))
	i = BigFromUint256(w)
	return
}

func SetStorageUint256(offset *Uint256, v *Uint256) {
	StorageCacheBytes32(unsafePtr(&offset), unsafePtr(&v))
}
