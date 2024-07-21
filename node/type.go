package node

import (
	"net"
	"sync"
)

const (
	LEADER = iota
	CANDIDATE
	FOLLOWER
)

type NodesPool struct {
	sync.Mutex
	Conns map[string]net.Conn
	Count int64
}

func (pool *NodesPool) Add(address string, conn net.Conn) {
	pool.Lock()
	defer pool.Unlock()
	pool.Conns[address] = conn
	pool.Count++
}

func (pool *NodesPool) Get(address string) net.Conn {
	pool.Lock()
	defer pool.Unlock()
	if pool.Count == 0 {
		return nil
	}
	return pool.Conns[address]
}

func (pool *NodesPool) Remove(address string, conn net.Conn) {
	pool.Lock()
	defer pool.Unlock()
	delete(pool.Conns, address)
	pool.Count--
}

type Node struct {
	nodeID      int64
	status      int64
	currentTerm int64
	voteFor     int64
	ALLNode     NodesPool
}

func NewNode() *Node {
	return &Node{
		nodeID:      0,
		status:      CANDIDATE,
		currentTerm: 0,
		voteFor:     0,
		ALLNode:     NodesPool{},
	}
}
