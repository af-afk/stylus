package stylus

import "math/big"

//go:wasmimport vm_hooks storage_load_bytes32
func storageLoadBytes32(key, dest int32)

//go:wasmimport vm_hooks storage_cache_bytes32
func storageCacheBytes32(key, src int32)

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
	SetStorageBig(&s.o, &b)
}

// LoadStorageBig from memory, using a fixed-size storage for it.
func LoadStorageBig(offset *Uint256) (i *big.Int) {
	var w [32]byte
	b := &w
	storageLoadBytes32(unsafePtr(&offset), unsafePtr(&b))
	i = BigFromUint256(w)
	return
}

func SetStorageBig(offset *Uint256, v *Uint256) {
	storageCacheBytes32(unsafePtr(&offset), unsafePtr(&v))
}
