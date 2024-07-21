package matchIndex

import "github.com/lvkeliang/Graft/model"

type MatchIndex struct {
	hashMap model.MessageHashMap
}

func NewMatchIndex() *MatchIndex {
	return new(MatchIndex)
}

func (index MatchIndex) add(nodeId int64, entryIdx int64) {
	index.hashMap.Add(nodeId, entryIdx)
}

func (index MatchIndex) del(nodeId int64, entryIdx int64) {
	index.hashMap.Del(nodeId, entryIdx)
}
