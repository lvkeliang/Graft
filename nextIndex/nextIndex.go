package nextIndex

import "github.com/lvkeliang/Graft/model"

type NextIndex struct {
	hashMap model.MessageHashMap
}

func NewNextIndex() *NextIndex {
	return new(NextIndex)
}

func (index NextIndex) add(nodeId int64, entryIdx int64) {
	index.hashMap.Add(nodeId, entryIdx)
}

func (index NextIndex) del(nodeId int64, entryIdx int64) {
	index.hashMap.Del(nodeId, entryIdx)
}
