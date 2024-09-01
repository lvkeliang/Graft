package Graft

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lvkeliang/Graft/LogEntry"
	"github.com/lvkeliang/Graft/matchIndex"
	"github.com/lvkeliang/Graft/nextIndex"
	"github.com/lvkeliang/Graft/stateMachine"
	"log"
	"math/rand"
	"net"
	"os"
	"sort"
	"sync"
	"time"
)

type StateOfNode int64

const (
	LEADER StateOfNode = iota
	CANDIDATE
	FOLLOWER
)

type nodeConn struct {
	RPCListenAddress string
	Conn             net.Conn
}

func NewNodeConn(RPCListenAddress string, conn net.Conn) nodeConn {
	return nodeConn{RPCListenAddress: RPCListenAddress, Conn: conn}
}

type NodesPool struct {
	sync.Mutex
	Conns map[string]nodeConn
	Count int64
}

func (pool *NodesPool) Add(RPCListenAddress string, conn net.Conn) {
	pool.Lock()
	defer pool.Unlock()

	pool.Conns[conn.RemoteAddr().String()] = NewNodeConn(RPCListenAddress, conn)
	pool.Count++
}

func (pool *NodesPool) Get(address string) net.Conn {
	pool.Lock()
	defer pool.Unlock()
	if pool.Count == 0 {
		return nil
	}
	return pool.Conns[address].Conn
}

func (pool *NodesPool) GetRPCListenAddress(address string) string {
	pool.Lock()
	defer pool.Unlock()
	if pool.Count == 0 {
		return ""
	}
	return pool.Conns[address].RPCListenAddress
}

func (pool *NodesPool) Remove(conn net.Conn) {
	pool.Lock()
	defer pool.Unlock()
	delete(pool.Conns, conn.RemoteAddr().String())
	pool.Count--
}

func (pool *NodesPool) GetALLRPCListenAddresses() []string {
	pool.Lock()
	defer pool.Unlock()
	if pool.Count == 0 {
		return nil
	}

	var addresses []string

	for _, connAddr := range pool.Conns {
		addresses = append(addresses, connAddr.RPCListenAddress)
	}

	return addresses
}

type Node struct {
	Mu            sync.Mutex
	RPCListenPort string
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
	StateMachine  stateMachine.StateMachine
}

type PersistentState struct {
	CurrentTerm int64
	VoteFor     string
	CommitIndex int64
	LastApplied int64
}

func NewNode(RPCListenPort string, stateFilePath string, logFilePath string, YourStateMachine stateMachine.StateMachine) *Node {
	logEnt, err := LogEntry.NewLog(logFilePath)
	if err != nil {
		log.Printf("[NewNode] init logEntry failed: %v\n", err)
		return nil
	}

	node := &Node{
		RPCListenPort: RPCListenPort,
		Status:        FOLLOWER,
		CurrentTerm:   0,
		VoteFor:       "",
		Log:           logEnt,
		CommitIndex:   -1,
		LastApplied:   -1,
		NextIndex:     nextIndex.NewNextIndex(),
		MatchIndex:    matchIndex.NewMatchIndex(),
		ElectionTimer: time.NewTimer(RandomElectionTimeout()),
		ALLNode: &NodesPool{
			Conns: make(map[string]nodeConn),
		},
		filePath:     stateFilePath,
		StateMachine: YourStateMachine, // 初始化状态机
	}

	err = node.load()
	if err != nil {
		log.Println("[NewNode] failed to load persistent state:", err)
	}

	return node
}

// ResetElectionTimer resets the election timer.
func (node *Node) ResetElectionTimer() {
	node.ElectionTimer.Stop()
	node.ElectionTimer.Reset(RandomElectionTimeout())
}

// RandomElectionTimeout generates a random election timeout duration.
func RandomElectionTimeout() time.Duration {
	return time.Duration(5000+rand.Intn(1500)) * time.Millisecond
}

func (node *Node) AddNode(RPCListenPort string, conn net.Conn) {
	conn.RemoteAddr().String()
	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return
	}

	node.ALLNode.Add(host+":"+RPCListenPort, conn)
	node.MatchIndex.Update(conn.RemoteAddr().String(), -1)
	node.NextIndex.Update(conn.RemoteAddr().String(), -1)
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

// UpdateCommitIndex updates the commitIndex based on the matchIndex values.
func (node *Node) UpdateCommitIndex() {

	matchIndexes := node.MatchIndex.GetAll()

	// Create a slice of all match indexes
	var indexes []int64
	for _, idx := range matchIndexes {
		indexes = append(indexes, idx)
	}

	// Sort the indexes slice
	sort.Slice(indexes, func(i, j int) bool {
		return indexes[i] < indexes[j]
	})

	// Find the majority match index (the middle value in the sorted list)
	quorumIndex := indexes[len(indexes)/2]

	fmt.Println("quorumIndex :", quorumIndex)

	if quorumIndex < 0 {
		return
	}

	// Check if the quorumIndex can be committed
	quorumEntry, err := node.Log.Get(quorumIndex)
	if err != nil {
		log.Println("[UpdateCommitIndex] failed to find quorumLogEntry: ", quorumEntry)
		return
	}

	if quorumIndex > node.CommitIndex && quorumEntry.Term == node.CurrentTerm {
		node.CommitIndex = quorumIndex
		// Apply the committed entries to the state machine
		node.applyLogEntries()
	}

	node.persist()

}

func (node *Node) applyLogEntries() {
	for node.LastApplied < node.CommitIndex {
		node.LastApplied++

		log.Println("LastApplied: ", node.LastApplied)
		entry, err := node.Log.Get(node.LastApplied)
		if err != nil {
			log.Printf("[applyLogEntries] failed to get log entry: %v | with CommitIndex : %v | with LastApplied : %v\n", err, node.CommitIndex, node.LastApplied)
			return
		}

		log.Printf("--------------------1-----------------------")
		result := node.StateMachine.ToApply(entry.Command)
		fmt.Println("--------------------2-----------------------")
		log.Printf("[applyLogEntries] Applied command '%s' with result '%s'", entry.Command, result)
		node.persist()
	}
}

// UpdateCommitIndexByLeader updates the commitIndex based on the matchIndex values.
func (node *Node) UpdateCommitIndexByLeader(index int64) {
	node.CommitIndex = index
	node.applyLogEntries()
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
