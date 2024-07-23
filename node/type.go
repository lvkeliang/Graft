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

func (pool *NodesPool) Add(conn net.Conn) {
	pool.Lock()
	defer pool.Unlock()
	pool.Conns[conn.RemoteAddr().String()] = conn
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

func (pool *NodesPool) Remove(conn net.Conn) {
	pool.Lock()
	defer pool.Unlock()
	delete(pool.Conns, conn.RemoteAddr().String())
	pool.Count--
}

type Node struct {
	NodeID      int64
	Status      int64
	CurrentTerm int64
	VoteFor     string
	ALLNode     NodesPool
}

func NewNode() *Node {
	return &Node{
		NodeID:      0,
		Status:      CANDIDATE,
		CurrentTerm: 0,
		VoteFor:     "",
		ALLNode: NodesPool{
			Conns: make(map[string]net.Conn),
		},
	}
}
