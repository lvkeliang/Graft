package node

import (
	"errors"
	"github.com/lvkeliang/Graft/LogEntry"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
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
	Mu            sync.Mutex
	ID            string
	Status        StateOfNode
	CurrentTerm   int64
	VoteFor       string
	Log           *LogEntry.Log
	CommitIndex   int64
	LastApplied   int64
	NextIndex     []int64
	MatchIndex    []int64
	ElectionTimer *time.Timer
	ALLNode       *NodesPool
}

func NewNode() *Node {
	logEnt, err := LogEntry.NewLog("raft_log.json")
	if err != nil {
		log.Println("[NewNode] init logEntry failed")
		return nil
	}

	return &Node{
		ID:            "",
		Status:        FOLLOWER,
		CurrentTerm:   0,
		VoteFor:       "",
		Log:           logEnt,
		CommitIndex:   0,
		LastApplied:   0,
		NextIndex:     make([]int64, 0),
		MatchIndex:    make([]int64, 0),
		ElectionTimer: time.NewTimer(RandomElectionTimeout()),
		ALLNode: &NodesPool{
			Conns: make(map[string]net.Conn),
		},
	}
}

// ResetElectionTimer resets the election timer.
func (r *Node) ResetElectionTimer() {
	r.ElectionTimer.Stop()
	r.ElectionTimer.Reset(RandomElectionTimeout())
}

// RandomElectionTimeout generates a random election timeout duration.
func RandomElectionTimeout() time.Duration {
	return time.Duration(1500+rand.Intn(1500)) * time.Millisecond
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
