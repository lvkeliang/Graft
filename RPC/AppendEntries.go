package RPC

import (
	"context"
	"fmt"
	"github.com/lvkeliang/Graft/node"
	"github.com/lvkeliang/Graft/protocol"
	"log"
	"net"
	"time"
)

// 用于将收到的leaderHeartbeat传递给vote相关
var leaderHeartbeat = make(chan int64)

func StartAppendEntries(ctx context.Context, myNode *node.Node) {
	ticker := time.NewTicker(5000 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if myNode.Status != node.LEADER {
				continue
			}
			myNode.ALLNode.Lock()
			for address, conn := range myNode.ALLNode.Conns {

				appendEntries := protocol.NewAppendEntries()
				appendEntries = &protocol.AppendEntries{
					Term:         myNode.CurrentTerm,
					LeaderID:     myNode.ID,
					LeaderCommit: myNode.CommitIndex,
				}

				// 使用 nextIndex 来选择 PrevLogIndex 和 PrevLogTerm
				nextLogIndex, ok := myNode.NextIndex.Get(address)
				prevLogIndex := nextLogIndex - 1

				if !ok {
					log.Println("[StartAppendEntries] nextIndex address not found")
					continue
				}

				if prevLogIndex >= 0 {
					prevLogEntry, err := myNode.Log.Get(prevLogIndex)
					if err != nil {
						log.Printf("[StartAppendEntries] get prevLogIndex log index out of range: %v\n", prevLogIndex)
						continue
					}
					appendEntries.PrevLogIndex = prevLogIndex
					appendEntries.PrevLogTerm = prevLogEntry.Term
				} else {
					appendEntries.PrevLogIndex = -1
					appendEntries.PrevLogTerm = -1
				}

				// 从 nextIndex 开始添加日志条目
				fmt.Println("myNode.Log.LastIndex(): ", myNode.Log.LastIndex())
				fmt.Println("nextLogIndex: ", nextLogIndex)
				if nextLogIndex == myNode.Log.LastIndex()+1 {
					// 如果follower已经有了完全的log，则发送nil
					appendEntries.Entries = nil
				} else if nextLogIndex == -1 {
					// 如果是初始化的nextLogIndex(标记为-1),则发送空的AppendEntries以重新收集各个节点的nextIndex,用以保证一致性
					// fmt.Println("Initial AppendEntries")
					appendEntries.Entries = nil
				} else {
					index, err := myNode.Log.GetAfterIndex(nextLogIndex)
					if err != nil {
						log.Printf("[StartAppendEntries] getAfterIndex index out of range: %v\n", nextLogIndex)
						continue
					}
					appendEntries.Entries = index
				}

				marshalAE, err := appendEntries.Marshal()
				fmt.Println(string(marshalAE))
				if err != nil {
					log.Println("[StartAppendEntries] appendEntries.Marshal failed")
					continue
				}

				_, err = conn.Write(marshalAE)
				if err != nil {
					log.Println("[StartAppendEntries] send marshalAE failed")
					continue
				}
			}
			myNode.ALLNode.Unlock()
		}

		ticker.Reset(5000 * time.Millisecond)
	}
}

func AppendEntriesHandle(conn net.Conn, myNode *node.Node, length int) {
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
	leaderHeartbeat <- appendEntries.Term
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

		fmt.Println(true)
		if appendEntries.Entries != nil {
			// 添加日志条目
			myNode.Log.AppendEntries(appendEntries.Entries)

			// 更新 commitIndex 和 lastApplied
			if appendEntries.LeaderCommit > myNode.CommitIndex {
				myNode.UpdateCommitIndex(appendEntries.LeaderCommit)
				if myNode.CommitIndex > myNode.Log.LastIndex() {
					myNode.UpdateCommitIndex(myNode.Log.LastIndex())
				}
				myNode.UpdateLastApplied(myNode.CommitIndex)
			}
		}

		res.LastIndex = myNode.Log.LastIndex()
	} else {
		res.LastIndex = myNode.Log.LastIndex()
	}

	res.Term = myNode.CurrentTerm

	fmt.Println(myNode.Log.Entries)
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
}

func AppendEntriesResultHandle(conn net.Conn, myNode *node.Node, length int) {

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

	// 使用 nextIndex 来选择 PrevLogIndex 和 PrevLogTerm
	nextLogIndex, ok := myNode.NextIndex.Get(conn.RemoteAddr().String())

	prevLogIndex := nextLogIndex - 1
	if !ok {
		log.Println("[AppendEntriesResultHandle] address not found")
		return
	}

	// 处理响应
	if appendEntriesResult.Success {
		myNode.MatchIndex.Update(conn.RemoteAddr().String(), prevLogIndex)
		myNode.NextIndex.Update(conn.RemoteAddr().String(), appendEntriesResult.LastIndex+1)
		fmt.Printf("update nextLogIndex: %v\n", myNode.Log.LastIndex()+1)
	} else {
		fmt.Printf("Decrement\n")
		myNode.NextIndex.Update(conn.RemoteAddr().String(), appendEntriesResult.LastIndex+1)
		if idx, ok := myNode.NextIndex.Get(conn.RemoteAddr().String()); ok && idx < 0 {
			myNode.NextIndex.Update(conn.RemoteAddr().String(), 0)
		}
	}

}
