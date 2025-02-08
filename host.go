package stylus

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
