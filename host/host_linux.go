//go:build !tinygo

package host

func PayForMemoryGrow(pages uint32) {}

func ReadArgs(dest int32) {}

func WriteResult(data int32, len int32) {}

func StorageFlushCache(clear bool) {}

func StorageLoadBytes32(key, dest int32) {}

func StorageCacheBytes32(key, src int32) {}
