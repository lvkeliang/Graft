package nextIndex

import (
	"sync"
)

type NextIndex struct {
	hashMap map[string]int64
	mu      sync.RWMutex
}

// NewNextIndex creates a new NextIndex instance.
func NewNextIndex() *NextIndex {
	return &NextIndex{
		hashMap: make(map[string]int64),
	}
}

// Update sets the next index for a given node.
func (index *NextIndex) Update(nodeId string, entryIdx int64) {
	index.mu.Lock()
	defer index.mu.Unlock()
	index.hashMap[nodeId] = entryIdx
}

// Del removes the next index for a given node.
func (index *NextIndex) Del(nodeId string) {
	index.mu.Lock()
	defer index.mu.Unlock()
	delete(index.hashMap, nodeId)
}

// Get retrieves the next index for a given node.
func (index *NextIndex) Get(nodeId string) (int64, bool) {
	index.mu.RLock()
	defer index.mu.RUnlock()
	val, ok := index.hashMap[nodeId]
	return val, ok
}

// GetAll returns a copy of the entire hashMap.
func (index *NextIndex) GetAll() map[string]int64 {
	index.mu.RLock()
	defer index.mu.RUnlock()
	copyMap := make(map[string]int64, len(index.hashMap))
	for k, v := range index.hashMap {
		copyMap[k] = v
	}
	return copyMap
}

// Decrement decreases the next index for a given node by one.
func (index *NextIndex) Decrement(nodeId string) {
	index.mu.Lock()
	defer index.mu.Unlock()
	if idx, ok := index.hashMap[nodeId]; ok {
		index.hashMap[nodeId] = idx - 1
	}
}

func (index *NextIndex) Reset() {
	index.mu.Lock()
	defer index.mu.Unlock()
	for nodeId := range index.hashMap {
		index.hashMap[nodeId] = -1
	}
}

// MajorityAtOrAbove checks if a majority of nodes have their next index at or above the given entryIdx.
func (index *NextIndex) MajorityAtOrAbove(entryIdx int64) bool {
	index.mu.RLock()
	defer index.mu.RUnlock()

	count := 0
	for _, idx := range index.hashMap {
		if idx >= entryIdx {
			count++
		}
	}

	// Check if the count is greater than half of the total nodes.
	return count > len(index.hashMap)/2
}
