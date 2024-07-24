package node

import (
	"errors"
	"net"
	"sync"
)

type StateOfNode int64

const (
	LEADER StateOfNode = iota
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
	Status      StateOfNode
	CurrentTerm int64
	VoteFor     string
	ALLNode     NodesPool
}

func NewNode() *Node {
	return &Node{
		NodeID:      0,
		Status:      FOLLOWER,
		CurrentTerm: 0,
		VoteFor:     "",
		ALLNode: NodesPool{
			Conns: make(map[string]net.Conn),
		},
	}
}

func (node *Node) SetVoteFor(conn net.Conn) error {
	if node.VoteFor != "" {
		return errors.New("this term has already voted, if want to vote please update term")
	}
	node.VoteFor = conn.RemoteAddr().String()
	return nil
}

func (node *Node) UpdateTerm(term int64) {
	node.CurrentTerm = term
	node.VoteFor = ""
}

func (node *Node) TermAddOne() {
	node.CurrentTerm++
	node.VoteFor = ""
}
