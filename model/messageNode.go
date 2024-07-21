package model

type MessageNode struct {
	nodeId   int64
	entryIdx int64
}

type MessageHashMap map[MessageNode]bool

func (hashMap MessageHashMap) Add(nodeId int64, entryIdx int64) {
	hashMap[MessageNode{
		nodeId:   nodeId,
		entryIdx: entryIdx,
	}] = true
}

func (hashMap MessageHashMap) Del(nodeId int64, entryIdx int64) {
	delete(hashMap, MessageNode{
		nodeId:   nodeId,
		entryIdx: entryIdx,
	})
}
