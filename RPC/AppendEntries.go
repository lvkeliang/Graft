package RPC

import (
	"context"
	"fmt"
	"github.com/lvkeliang/Graft/LogEntry"
	"github.com/lvkeliang/Graft/node"
	"github.com/lvkeliang/Graft/protocol"
	"log"
	"net"
	"time"
)

// 用于将收到的leaderHeartbeat传递给vote相关
var leaderHeartbeat = make(chan int64)

func StartAppendEntries(ctx context.Context, myNode *node.Node, logEnt *LogEntry.LogEntry) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	appendEntries := protocol.NewAppendEntries()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if myNode.Status != node.LEADER {
				continue
			}
			myNode.ALLNode.Lock()
			for ip, conn := range myNode.ALLNode.Conns {
				//TODO: 创建一个AppendEntries协议，发送heartbeat以及log

				appendEntries.Term = myNode.CurrentTerm
				appendEntries.LeaderIP = myNode.VoteFor
				appendEntries.PrevLogIndex = logEnt.PrevLog[ip].Idx
				appendEntries.LeaderCommit = logEnt.CommitIndex
				appendEntries.PrevLogTerm = logEnt.PrevLog[ip].Term

				if len(logEnt.LogItem) > 1 {
					appendEntries.Entries = append([][]string{logEnt.LogItem[logEnt.PrevLog[ip].Term][logEnt.PrevLog[ip].Idx:]}, logEnt.LogItem[logEnt.PrevLog[ip].Term+1:][:]...)
				} else {
					appendEntries.Entries = logEnt.LogItem[logEnt.PrevLog[ip].Term:][logEnt.PrevLog[ip].Idx:]

				}

				marshalAE, err := appendEntries.Marshal()
				if err != nil {
					log.Println("[StartAppendEntries] appendEntries.Marshal failed")
					continue
				}

				_, err = conn.Write(marshalAE)
				if err != nil {
					log.Println("[StartAppendEntries] send marshalAE failed")
					return
				}
			}
			myNode.ALLNode.Unlock()
		}
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
	result := protocol.NewAppendEntries()
	resup := buf[:n]
	err = result.UNMarshal(resup)
	if err != nil {
		log.Println("[AppendEntriesHandle] Error unmarshaling data:", err)
		log.Println(string(resup))
		return
	}

	// 将收到的leaderHeartbeat传递给vote相关
	// 更新自己的term到leader的term
	leaderHeartbeat <- result.Term
	//fmt.Printf("[AppendEntriesHandle] leader term: %v\n", result.Term)

	res := protocol.NewAppendEntriesResult()

	res.Success = true
	res.Term = myNode.CurrentTerm
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

func AppendEntriesResultHandle(conn net.Conn, length int) {

	// Read data from the connection
	buf := make([]byte, length)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("[AppendEntriesResultHandle] Error reading data:", err)
		return
	}

	// Parse received data into AppendEntriesResult
	result := protocol.NewAppendEntriesResult()
	err = result.UNMarshal(buf[:n])
	if err != nil {
		log.Println("[AppendEntriesResultHandle] Error unmarshaling data:", err)
		return
	}

	// Process the result (e.g., update leader's term, handle success/failure)
	if result.Success {
		// fmt.Println("[AppendEntriesResultHandle] Received successful AppendEntriesResult")
		// Update leader's term if needed
	} else {
		fmt.Println("[AppendEntriesResultHandle] Received failed AppendEntriesResult")
		// Handle failure case
	}

}
