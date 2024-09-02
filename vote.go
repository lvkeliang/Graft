package Graft

import (
	"context"
	"fmt"
	"github.com/lvkeliang/Graft/protocol"
	"log"
	"net"
)

// 创建一个用于收集投票结果的通道
var voteAccept = make(chan bool)

type voteRequest struct {
	address string
	Term    int64
}

// 创建一个用于作为follower接收到竞选请求时，抑制成为candidate的通道
var receivedVoteRequest = make(chan voteRequest)

func StartElection(ctx context.Context, myNode *Node) {
	var cancel context.CancelFunc
	myNode.ResetElectionTimer()

	for {
		select {
		case <-ctx.Done():
			if cancel != nil {
				cancel()
			}
			return
		case <-myNode.ElectionTimer.C:
			if cancel != nil {
				cancel()
			}
			switch myNode.Status {
			case FOLLOWER:
				myNode.Status = CANDIDATE
				myNode.TermAddOne()
				//fmt.Printf("节点转为 Candidate, 当前term: %v\n", myNode.CurrentTerm)
				myNode.ResetElectionTimer()

				var timerCtx context.Context
				timerCtx, cancel = context.WithCancel(context.Background())
				go collectVoteResults(timerCtx, myNode)
				RequestVote(myNode)
			case CANDIDATE:
				myNode.ResetElectionTimer()
				if myNode.ALLNode.Count > 0 {
					myNode.TermAddOne()
					var timerCtx context.Context
					timerCtx, cancel = context.WithCancel(context.Background())
					go collectVoteResults(timerCtx, myNode)
					RequestVote(myNode)
				}
			case LEADER:
				myNode.ResetElectionTimer()
			}
		case heartbeat := <-leaderHeartbeatChan:
			if myNode.Status == LEADER || myNode.Status == CANDIDATE {
				if myNode.CurrentTerm < heartbeat.Term {
					myNode.Status = FOLLOWER
					myNode.CurrentLeader = heartbeat.address // 更新Leader
					fmt.Printf("节点转为 Follower, 当前term: %v\n", myNode.CurrentTerm)
				}
			}
			if myNode.CurrentTerm <= heartbeat.Term || myNode.CurrentLeader != heartbeat.address {
				myNode.UpdateTerm(heartbeat.Term)
				myNode.CurrentLeader = heartbeat.address // 更新Leader
			}
			myNode.ResetElectionTimer()
		case candidateVoteRequest := <-receivedVoteRequest:
			if myNode.Status == FOLLOWER {
				myNode.ResetElectionTimer()
			} else if myNode.Status == CANDIDATE && candidateVoteRequest.Term > myNode.CurrentTerm {
				myNode.ResetElectionTimer()
				myNode.Status = FOLLOWER
				myNode.CurrentLeader = candidateVoteRequest.address // 更新Leader
				//fmt.Printf("节点转为 Follower, 当前term: %v\n", myNode.CurrentTerm)
			} else if myNode.Status == LEADER && candidateVoteRequest.Term > myNode.CurrentTerm {
				myNode.ResetElectionTimer()
				myNode.Status = FOLLOWER
				myNode.CurrentLeader = candidateVoteRequest.address // 更新Leader
				//fmt.Printf("节点转为 Follower, 当前term: %v\n", myNode.CurrentTerm)
			}
		}
	}
}

// 启动收集投票结果的协程
func collectVoteResults(ctx context.Context, myNode *Node) {
	var votesReceived int64

	for {
		select {
		case <-ctx.Done():
			// 上下文取消，停止收集投票结果
			return
		case <-voteAccept:
			votesReceived++
			if votesReceived > myNode.ALLNode.Count/2 {
				// 获得大于一半的节点的投票同意，晋升为 Leader
				myNode.Status = LEADER
				myNode.CurrentLeader = "self" // 自己成为Leader
				fmt.Printf("节点晋升为 Leader, 当前term: %v\n", myNode.CurrentTerm)

				//转为Leader时将nextIndex重置,由AppendEntries发出空的RPC请求以重新收集各个节点的nextIndex,用以保证一致性
				myNode.NextIndex.Reset()

				return
			}

		}
	}
}

func RequestVote(myNode *Node) {

	requeatVote := protocol.NewRequestVote()
	requeatVote.Term = myNode.CurrentTerm
	//requeatVote.CandidateID

	lastLogEntry := myNode.Log.GetLastEntry()
	if lastLogEntry != nil {
		requeatVote.LastLogIndex = lastLogEntry.Index
		requeatVote.LastLogTerm = lastLogEntry.Term
	} else {
		requeatVote.LastLogIndex = -1
		requeatVote.LastLogTerm = -1
	}

	marshalRV, err := requeatVote.Marshal()
	if err != nil {
		log.Println("[RequestVote] marshal requestVote failed")
		return
	}

	for _, nodeConn := range myNode.ALLNode.Conns {
		conn := nodeConn.Conn
		_, err = conn.Write(marshalRV)
		if err != nil {
			log.Println("[RequestVote] send requestVot failed")
			return
		}
	}

}

func RequestVoteHandle(conn net.Conn, myNode *Node, length int) {
	buf := make([]byte, length)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("[RequestVoteHandle] Error reading data:", err)
		return
	}

	request := protocol.NewRequestVote()
	resup := buf[:n]
	err = request.UNMarshal(resup)
	if err != nil {
		log.Println("[RequestVoteHandle] Error unmarshaling data:", err)
		//log.Println(string(resup))
		return
	}

	res := protocol.NewRequestVoteResult()
	res.Term = myNode.CurrentTerm

	if myNode.Status == LEADER {
		res.VoteGranted = false
		return
	}

	if request.Term >= myNode.CurrentTerm {
		lastLogEntry := myNode.Log.GetLastEntry()
		var lastTerm int64 = -1
		var lastIndex int64 = -1
		if lastLogEntry != nil {
			lastTerm = lastLogEntry.Term
			lastIndex = lastLogEntry.Index
		}

		if request.LastLogTerm > lastTerm || (request.LastLogTerm == lastTerm && request.LastLogIndex >= lastIndex) {
			res.VoteGranted = true
			myNode.CurrentLeader = conn.RemoteAddr().String() // 更新Leader
			//fmt.Printf("VoteGranted to : %v\n", myNode.CurrentLeader)

			err = myNode.SetVoteFor(conn)
			if err != nil {
				log.Println("[RequestVoteHandle] Error has already voted:", err)
				return
			}
		}
	}

	marshalRes, err := res.Marshal()
	if err != nil {
		log.Println("[RequestVoteHandle] Error marshaling res:", err)
		return
	}
	_, err = conn.Write(marshalRes)
	if err != nil {
		log.Println("[RequestVoteHandle] send marshalRes failed:", err)
		return
	}

	if res.VoteGranted == true {
		receivedVoteRequest <- voteRequest{
			address: conn.RemoteAddr().String(),
			Term:    res.Term,
		}
	}
}

func RequestVoteResultHandle(conn net.Conn, myNode *Node, length int) {
	// Read data from the connection
	buf := make([]byte, length)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("[RequestVoteResultHandle] Error reading data:", err)
		return
	}

	// Parse received data into AppendEntriesResult
	result := protocol.NewRequestVoteResult()
	err = result.UNMarshal(buf[:n])
	if err != nil {
		log.Println("[RequestVoteResultHandle] Error unmarshaling data:", err)
		return
	}

	// Process the result (e.g., update leader's term, handle success/failure)
	if result.VoteGranted {
		//fmt.Println("[RequestVoteResultHandle] Received successful RequestVoteResult")
		voteAccept <- true

	} else {
		//fmt.Println("[RequestVoteResultHandle] Received failed RequestVoteResult")
		if result.Term > myNode.CurrentTerm {
			// 存在term比自己大的节点
			// 放弃竞选
			myNode.Status = FOLLOWER
			//fmt.Printf("节点转为 Follower, 当前term: %v\n", myNode.CurrentTerm)
			return
		}

	}
}
