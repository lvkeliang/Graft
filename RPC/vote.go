package RPC

import (
	"context"
	"fmt"
	"github.com/lvkeliang/Graft/LogEntry"
	"github.com/lvkeliang/Graft/node"
	"github.com/lvkeliang/Graft/protocol"
	"log"
	"math/rand"
	"net"
	"time"
)

// 创建一个用于收集投票结果的通道
var voteAccept = make(chan bool)

// 创建一个用于作为follower接收到竞选请求时，抑制成为candidate的通道
var receivedVoteRequest = make(chan bool)

// 生成介于 min 和 max 微秒之间的随机持续时间的辅助函数
func randomDuration(min, max int) time.Duration {
	return time.Duration(rand.Intn(max-min+1)+min) * time.Microsecond * 1000
}

// StartElection 启动计时和发起选举请求的协程
func StartElection(ctx context.Context, myNode *node.Node, logEnt *LogEntry.LogEntry) {
	electionTimer := time.NewTimer(randomDuration(1500, 2000)) // 5s 到 6s

	// 用于取消选票的进程
	var cancel context.CancelFunc

	for {
		select {
		case <-ctx.Done():
			if cancel != nil {
				cancel()
			}

			// 上下文取消，停止选举过程
			return
		case <-electionTimer.C:
			// 取消上一个收集选票进程
			if cancel != nil {
				cancel()
			}
			switch myNode.Status {
			case node.FOLLOWER:
				// 转换为 Candidate
				myNode.Status = node.CANDIDATE
				// term+1
				myNode.TermAddOne()
				fmt.Printf("节点转为 Candidate, 当前term: %v\n", myNode.CurrentTerm)

				// 重置选举计时器
				electionTimer.Reset(randomDuration(1500, 2000))

				// 创建一个上下文
				var timerCtx context.Context
				timerCtx, cancel = context.WithCancel(context.Background())
				// 发送选举请求并收集选票
				go collectVoteResults(timerCtx, myNode)
				RequestVote(myNode, logEnt)
			case node.CANDIDATE:
				//该term内没有选举出leader
				// 重置选举计时器
				electionTimer.Reset(randomDuration(1500, 2000))

				// 转换为 FOLLOWER
				myNode.Status = node.CANDIDATE

				var timerCtx context.Context
				timerCtx, cancel = context.WithCancel(context.Background())
				go collectVoteResults(timerCtx, myNode)
				RequestVote(myNode, logEnt)

			case node.LEADER:
				electionTimer.Reset(randomDuration(1500, 2000))
			}
		case leaderTerm := <-leaderHeartbeat:
			// 已经收到了别的leader的心跳
			// 更新term为ter最新的leader的term
			if myNode.Status == node.LEADER {
				if myNode.CurrentTerm < leaderTerm {
					myNode.UpdateTerm(leaderTerm)
					//将自己的state置为follower
					myNode.Status = node.FOLLOWER
				}
			} else if myNode.Status == node.CANDIDATE {
				myNode.UpdateTerm(leaderTerm)
				//将自己的state置为follower
				myNode.Status = node.FOLLOWER
			} else if myNode.Status == node.FOLLOWER {
				myNode.UpdateTerm(leaderTerm)
			}

			// 重置选举计时器
			electionTimer.Reset(randomDuration(1500, 2000))
		case <-receivedVoteRequest:
			// 收到Vote请求后抑制成为candidate
			if myNode.Status == node.FOLLOWER {
				// 重置选举计时器
				electionTimer.Reset(randomDuration(1500, 2000))
			}
		}
	}
}

// 启动收集投票结果的协程
func collectVoteResults(ctx context.Context, myNode *node.Node) {
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
				myNode.Status = node.LEADER
				fmt.Printf("节点晋升为 Leader, 当前term: %v\n", myNode.CurrentTerm)
				return
			}

		}
	}
}

func RequestVote(myNode *node.Node, logEnt *LogEntry.LogEntry) {

	requeatVote := protocol.NewRequestVote()
	requeatVote.Term = myNode.CurrentTerm
	//requeatVote.CandidateIP
	requeatVote.LastLogTerm = logEnt.LastLog.Term
	requeatVote.LastLogIndex = logEnt.LastLog.Idx

	marshalRV, err := requeatVote.Marshal()
	if err != nil {
		log.Println("[RequestVote] marshal requestVote failed")
		return
	}

	for _, conn := range myNode.ALLNode.Conns {
		_, err = conn.Write(marshalRV)
		if err != nil {
			log.Println("[RequestVote] send requestVot failed")
			return
		}
	}

}

func RequestVoteHandle(conn net.Conn, myNode *node.Node, logEnt *LogEntry.LogEntry, length int) {

	// Read data from the connection
	buf := make([]byte, length)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("[RequestVoteHandle] Error reading data:", err)
		return
	}

	// Parse received data into AppendEntriesResult
	request := protocol.NewRequestVote()
	resup := buf[:n]
	err = request.UNMarshal(resup)
	if err != nil {
		log.Println("[RequestVoteHandle] Error unmarshaling data:", err)
		log.Println(string(resup))
		return
	}

	res := protocol.NewRequestVoteResult()
	res.Term = myNode.CurrentTerm

	//确保候选者的term要比本节点的大
	if myNode.Status == node.LEADER {
		res.VoteGranted = false
		return
	}

	if request.Term > myNode.CurrentTerm {
		// Check if the candidate's log is up-to-date
		if request.LastLogTerm > logEnt.LastLog.Term || (request.LastLogTerm == logEnt.LastLog.Term && request.LastLogIndex >= logEnt.LastLog.Idx) {
			res.VoteGranted = true
		}
	}

	marshalRes, err := res.Marshal()
	if err != nil {
		log.Println("[RequestVoteHandle] Error marshaling res:", err)
		return
	}
	_, err = conn.Write(marshalRes)
	if err != nil {
		log.Println("[RequestVoteHandle] send marshalRes faield:", err)
		return
	}

	// 抑制成为candidate
	receivedVoteRequest <- true

	// 更新信息
	myNode.UpdateTerm(request.Term)
	err = myNode.SetVoteFor(conn)
	if err != nil {
		log.Println("[RequestVoteHandle] Error has already voted:", err)
	}

}

func RequestVoteResultHandle(conn net.Conn, myNode *node.Node, length int) {
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
		fmt.Println("[RequestVoteResultHandle] Received successful RequestVoteResult")
		voteAccept <- true

	} else {
		fmt.Println("[RequestVoteResultHandle] Received failed RequestVoteResult")
		if result.Term > myNode.CurrentTerm {
			// 存在term比自己大的节点
			// 放弃竞选
			myNode.Status = node.FOLLOWER
			return
		}

	}
}
