package node

import (
	"encoding/json"
	"errors"
	"github.com/lvkeliang/Graft/LogEntry"
	"github.com/lvkeliang/Graft/matchIndex"
	"github.com/lvkeliang/Graft/nextIndex"
	"log"
	"math/rand"
	"net"
	"os"
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
	NextIndex     *nextIndex.NextIndex
	MatchIndex    *matchIndex.MatchIndex
	ElectionTimer *time.Timer
	ALLNode       *NodesPool
	filePath      string
}

type PersistentState struct {
	CurrentTerm int64
	VoteFor     string
	CommitIndex int64
	LastApplied int64
}

func NewNode(id string, stateFilePath string, logFilePath string) *Node {
	logEnt, err := LogEntry.NewLog(logFilePath)
	if err != nil {
		log.Printf("[NewNode] init logEntry failed: %v\n", err)
		return nil
	}

	node := &Node{
		ID:            id,
		Status:        FOLLOWER,
		CurrentTerm:   0,
		VoteFor:       "",
		Log:           logEnt,
		CommitIndex:   0,
		LastApplied:   0,
		NextIndex:     nextIndex.NewNextIndex(),
		MatchIndex:    matchIndex.NewMatchIndex(),
		ElectionTimer: time.NewTimer(RandomElectionTimeout()),
		ALLNode: &NodesPool{
			Conns: make(map[string]net.Conn),
		},
		filePath: stateFilePath,
	}

	err = node.load()
	if err != nil {
		log.Println("[NewNode] failed to load persistent state:", err)
	}

	return node
}

// ResetElectionTimer resets the election timer.
func (r *Node) ResetElectionTimer() {
	r.ElectionTimer.Stop()
	r.ElectionTimer.Reset(RandomElectionTimeout())
}

// RandomElectionTimeout generates a random election timeout duration.
func RandomElectionTimeout() time.Duration {
	return time.Duration(5000+rand.Intn(1500)) * time.Millisecond
}

func (node *Node) AddNode(conn net.Conn) {
	node.ALLNode.Add(conn)
	node.MatchIndex.Update(conn.RemoteAddr().String(), 0)
	node.NextIndex.Update(conn.RemoteAddr().String(), 0)
}

func (node *Node) RemoveNode(conn net.Conn) {
	node.ALLNode.Remove(conn)
	node.MatchIndex.Del(conn.RemoteAddr().String())
	node.NextIndex.Del(conn.RemoteAddr().String())
}

func (node *Node) SetVoteFor(conn net.Conn) error {
	if node.VoteFor != "" {
		return errors.New("this term has already voted, if want to vote please update term")
	}
	node.VoteFor = conn.RemoteAddr().String()
	node.persist()
	return nil
}

func (node *Node) UpdateTerm(term int64) {
	node.CurrentTerm = term
	node.VoteFor = ""
	node.persist()
}

func (node *Node) TermAddOne() {
	node.CurrentTerm++
	node.VoteFor = ""
	node.persist()
}

func (node *Node) UpdateCommitIndex(index int64) {
	node.CommitIndex = index
	node.persist()
}

func (node *Node) UpdateLastApplied(index int64) {
	node.LastApplied = index
	node.persist()
}

func (node *Node) persist() {
	node.Mu.Lock()
	defer node.Mu.Unlock()

	state := PersistentState{
		CurrentTerm: node.CurrentTerm,
		VoteFor:     node.VoteFor,
		CommitIndex: node.CommitIndex,
		LastApplied: node.LastApplied,
	}

	file, err := os.OpenFile(node.filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Println("[persist] failed to open file:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(state)
	if err != nil {
		log.Println("[persist] failed to encode state:", err)
	}
}

func (node *Node) load() error {
	node.Mu.Lock()
	defer node.Mu.Unlock()

	file, err := os.OpenFile(node.filePath, os.O_RDONLY, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // It's ok if the file does not exist
		}
		return err
	}
	defer file.Close()

	var state PersistentState
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&state)
	if err != nil {
		return err
	}

	node.CurrentTerm = state.CurrentTerm
	node.VoteFor = state.VoteFor
	node.CommitIndex = state.CommitIndex
	node.LastApplied = state.LastApplied

	return nil
}
