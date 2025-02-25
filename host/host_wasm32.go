//go:build tinygo

// Tinygo won't include files for wasm that aren't suffixed with wasm32,
// so we can blanket ban anything other than tinygo here.

package host

//go:wasmimport vm_hooks pay_for_memory_grow
func PayForMemoryGrow(pages uint32)

//go:wasmimport vm_hooks read_args
func ReadArgs(dest int32)

//go:wasmimport vm_hooks write_result
func WriteResult(data int32, len int32)

//go:wasmimport vm_hooks storage_flush_cache
func StorageFlushCache(clear bool)

//go:wasmimport vm_hooks storage_load_bytes32
func StorageLoadBytes32(key, dest int32)

//go:wasmimport vm_hooks storage_cache_bytes32
func StorageCacheBytes32(key, src int32)

//go:wasmimport vm_hooks call_contract
func CallContract(contract, calldata, calldataLen, value int32, gas uint64, returndataLen int32)
