package Graft

import (
	"context"
	"github.com/lvkeliang/Graft/protocol"
	"log"
	"net"
)

type leaderHeartbeat struct {
	address string
	Term    int64
}

// 用于将收到的leaderHeartbeat传递给vote相关
var leaderHeartbeatChan = make(chan leaderHeartbeat)

func StartAppendEntries(ctx context.Context, myNode *Node) {
	defer myNode.AppendEntriesTimer.Stop()
	myNode.ResetAppendEntriesTimer()

	for {
		select {
		case <-ctx.Done():
			return
		case <-myNode.AppendEntriesTimer.C:
			if myNode.Status != LEADER {
				myNode.ResetAppendEntriesTimer()
				continue
			}
			myNode.ALLNode.Lock()
			for address, nodeConn := range myNode.ALLNode.Conns {
				conn := nodeConn.Conn

				appendEntries := protocol.NewAppendEntries()
				appendEntries = &protocol.AppendEntries{
					Term:         myNode.CurrentTerm,
					LeaderID:     myNode.RPCListenPort,
					LeaderCommit: myNode.CommitIndex,
				}

				// 使用 nextIndex 来选择 PrevLogIndex 和 PrevLogTerm
				nextLogIndex, ok := myNode.NextIndex.Get(address)
				prevLogIndex := nextLogIndex - 1

				if !ok {
					log.Println("[StartAppendEntries] nextIndex address not found")
					myNode.ResetAppendEntriesTimer()
					continue
				}

				if prevLogIndex >= 0 {
					prevLogEntry, err := myNode.Log.Get(prevLogIndex)
					if err != nil {
						log.Printf("[StartAppendEntries] get prevLogIndex log index out of range: %v\n", prevLogIndex)
						myNode.ResetAppendEntriesTimer()
						continue
					}
					appendEntries.PrevLogIndex = prevLogIndex
					appendEntries.PrevLogTerm = prevLogEntry.Term
				} else {
					appendEntries.PrevLogIndex = -1
					appendEntries.PrevLogTerm = -1
				}

				// 从 nextIndex 开始添加日志条目
				//fmt.Println("myNode.Log.LastIndex(): ", myNode.Log.LastIndex())
				//fmt.Println("nextLogIndex: ", nextLogIndex)
				if nextLogIndex == myNode.Log.LastIndex()+1 {
					// 如果follower已经有了完全的log，则发送nil
					appendEntries.Entries = nil
				} else if nextLogIndex == -1 {
					// 如果是初始化的nextLogIndex(标记为-1),则发送空的AppendEntries以重新收集各个节点的nextIndex,用以保证一致性
					// fmt.Println("Initial AppendEntries")
					appendEntries.Entries = nil
					myNode.NextIndex.Update(conn.RemoteAddr().String(), 0)
				} else {
					index, err := myNode.Log.GetAfterIndex(nextLogIndex)
					if err != nil {
						log.Printf("[StartAppendEntries] getAfterIndex index out of range: %v\n", nextLogIndex)
						myNode.ResetAppendEntriesTimer()
						continue
					}
					appendEntries.Entries = index

				}

				marshalAE, err := appendEntries.Marshal()
				// fmt.Println(string(marshalAE))
				if err != nil {
					log.Println("[StartAppendEntries] appendEntries.Marshal failed", err)
					myNode.ResetAppendEntriesTimer()
					continue
				}

				_, err = conn.Write(marshalAE)
				if err != nil {
					log.Println("[StartAppendEntries] send marshalAE failed: ", err)
					myNode.RemoveNode(conn)
					myNode.ResetAppendEntriesTimer()
					continue
				}

				if appendEntries.Entries != nil {
					myNode.NextIndex.Update(conn.RemoteAddr().String(), appendEntries.Entries[len(appendEntries.Entries)-1].Index+1)
					//fmt.Printf("update nextLogIndex: %v\n", appendEntries.Entries[len(appendEntries.Entries)-1].Index+1)
				}

				//fmt.Println("[AppendEntries] MatchIndex: ", myNode.MatchIndex.GetAll())
			}
			myNode.ALLNode.Unlock()
		}

		myNode.ResetAppendEntriesTimer()
	}
}

func AppendEntriesHandle(conn net.Conn, myNode *Node, length int) {
	// Read data from the connection
	buf := make([]byte, length)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("[AppendEntriesHandle] Error reading data:", err)
		return
	}

	// Parse received data into AppendEntriesResult
	appendEntries := protocol.NewAppendEntries()
	err = appendEntries.UNMarshal(buf[:n])
	if err != nil {
		log.Println("[AppendEntriesHandle] Error unmarshaling data:", err)
		return
	}

	// 将收到的leaderHeartbeat传递给vote相关
	// 更新自己的term到leader的term
	leaderHeartbeatChan <- leaderHeartbeat{
		address: conn.RemoteAddr().String(),
		Term:    appendEntries.Term,
	}
	//fmt.Printf("[AppendEntriesHandle] leader term: %v\n", appendEntries.Term)

	res := protocol.NewAppendEntriesResult()

	// 验证 PrevLogIndex 和 PrevLogTerm
	res.Success = false

	var prevTerm int64 = -1
	if appendEntries.PrevLogIndex >= 0 {
		prevLog, err := myNode.Log.Get(appendEntries.PrevLogIndex)
		if err != nil {
			log.Printf("[AppendEntriesHandle] index out of range: %v\n", appendEntries.PrevLogIndex)
		}

		prevTerm = prevLog.Term
	}

	if (appendEntries.PrevLogIndex == -1 && myNode.Log.LastIndex() == -1) || (appendEntries.PrevLogIndex == myNode.Log.LastIndex()) && (prevTerm == appendEntries.PrevLogTerm) {
		res.Success = true

		if appendEntries.Entries != nil {
			// 添加日志条目
			myNode.Log.AppendEntries(appendEntries.Entries)
			//fmt.Printf("[AppendEntriesHandle] appendEntries.LeaderCommit: %v | myNode.CommitIndex: %v | LastIndex: %v\n", appendEntries.LeaderCommit, myNode.CommitIndex, myNode.Log.LastIndex())
		}

		// 更新 commitIndex 和 lastApplied
		if appendEntries.LeaderCommit > myNode.CommitIndex {

			if appendEntries.LeaderCommit > myNode.Log.LastIndex() {
				myNode.UpdateCommitIndexByLeader(myNode.Log.LastIndex())
			} else {
				myNode.UpdateCommitIndexByLeader(appendEntries.LeaderCommit)
			}
		}

		res.LastIndex = myNode.Log.LastIndex()
	} else {
		res.LastIndex = myNode.Log.LastIndex()
	}

	res.Term = myNode.CurrentTerm

	//fmt.Println(myNode.Log.Entries)
	marshalRes, err := res.Marshal()
	if err != nil {
		log.Println("[AppendEntriesHandle] Error marshaling res:", err)
		return
	}

	_, err = conn.Write(marshalRes)
	if err != nil {
		log.Println("[AppendEntriesHandle] send marshalRes faield:", err)
		return
	}

	//fmt.Println("[AppendEntriesHandle] sent:", res)
}

func AppendEntriesResultHandle(conn net.Conn, myNode *Node, length int) {

	// Read data from the connection
	buf := make([]byte, length)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("[AppendEntriesResultHandle] Error reading data:", err)
		return
	}

	// Parse received data into AppendEntriesResult
	appendEntriesResult := protocol.NewAppendEntriesResult()
	err = appendEntriesResult.UNMarshal(buf[:n])
	if err != nil {
		log.Println("[AppendEntriesResultHandle] Error unmarshaling data:", err)
		return
	}

	// 处理响应
	if appendEntriesResult.Success {
		myNode.MatchIndex.Update(conn.RemoteAddr().String(), appendEntriesResult.LastIndex)
		myNode.UpdateCommitIndex()
	} else {

		myNode.NextIndex.Update(conn.RemoteAddr().String(), appendEntriesResult.LastIndex+1)
		if idx, ok := myNode.NextIndex.Get(conn.RemoteAddr().String()); ok && idx < 0 {
			myNode.NextIndex.Update(conn.RemoteAddr().String(), 0)
		}
	}

}
