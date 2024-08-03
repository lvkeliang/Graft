package matchIndex

import "sync"

type MatchIndex struct {
	hashMap map[string]int64
	mu      sync.RWMutex
}

// NewMatchIndex creates a new MatchIndex instance.
func NewMatchIndex() *MatchIndex {
	return &MatchIndex{
		hashMap: make(map[string]int64),
	}
}

// Update sets the match index for a given node.
func (index *MatchIndex) Update(nodeId string, entryIdx int64) {
	index.mu.Lock()
	defer index.mu.Unlock()
	index.hashMap[nodeId] = entryIdx
}

// Del removes the match index for a given node.
func (index *MatchIndex) Del(nodeId string) {
	index.mu.Lock()
	defer index.mu.Unlock()
	delete(index.hashMap, nodeId)
}

// Get retrieves the match index for a given node.
func (index *MatchIndex) Get(nodeId string) (int64, bool) {
	index.mu.RLock()
	defer index.mu.RUnlock()
	val, ok := index.hashMap[nodeId]
	return val, ok
}

// GetAll returns a copy of the entire hashMap.
func (index *MatchIndex) GetAll() map[string]int64 {
	index.mu.RLock()
	defer index.mu.RUnlock()
	copyMap := make(map[string]int64, len(index.hashMap))
	for k, v := range index.hashMap {
		copyMap[k] = v
	}
	return copyMap
}

// MajorityConfirmed checks if a majority of nodes have confirmed the given entryIdx.
func (index *MatchIndex) MajorityConfirmed(entryIdx int64) bool {
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
