package busconfig

import (
	"sync"
)

var cache *cacheModule

type cacheModule struct {
	mu             sync.Mutex
	rankTableCache map[string]string
}

func init() {
	cache = &cacheModule{
		mu:             sync.Mutex{},
		rankTableCache: make(map[string]string),
	}
}

// AddData add data
func AddData(jobId string, data string) {
	cache.mu.Lock()
	cache.rankTableCache[jobId] = data
	cache.mu.Unlock()
	dataChanged(jobId, data)
}

// GetData get cached by job id
func GetData(jobId string) string {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if table, ok := cache.rankTableCache[jobId]; ok {
		return table
	}
	return ""
}
