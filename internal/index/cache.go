package index

import "sync"

var linkCache = struct {
	mu    sync.RWMutex
	items map[string]BacklinkIndex
}{
	items: map[string]BacklinkIndex{},
}

func GetCached(vaultRoot string) (BacklinkIndex, bool) {
	linkCache.mu.RLock()
	defer linkCache.mu.RUnlock()
	idx, ok := linkCache.items[vaultRoot]
	return idx, ok
}

func SetCached(vaultRoot string, idx BacklinkIndex) {
	linkCache.mu.Lock()
	defer linkCache.mu.Unlock()
	linkCache.items[vaultRoot] = idx
}
