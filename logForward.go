package Graft

import (
	"errors"
	"github.com/lvkeliang/Graft/LogEntry"
	"github.com/lvkeliang/Graft/protocol"
	"log"
	"net"
)

func ForwardLogEntry(myNode *Node, command string) error {
	if myNode.CurrentLeader == "" {
		return errors.New("no leader available")
	}

	conn := myNode.ALLNode.Get(myNode.CurrentLeader)
	if conn == nil {
		return errors.New("failed to connect to leader")
	}

	// 准备要转发的日志条目
	logEntry := LogEntry.LogEntry{
		Term:    myNode.CurrentTerm,
		Command: command,
	}

	logForward := protocol.NewLogForward()
	logForward.Term = myNode.CurrentTerm
	logForward.NodeID = myNode.RPCListenPort
	logForward.Entries = []LogEntry.LogEntry{logEntry}

	marshalLF, err := logForward.Marshal()
	if err != nil {
		return err
	}

	_, err = conn.Write(marshalLF)
	if err != nil {
		myNode.RemoveNode(conn)
		return err
	}

	//log.Printf("[ForwardLogEntry] Successfully forwarded log entry to leader %s", myNode.CurrentLeader)
	return nil
}

func LogForwardHandle(conn net.Conn, myNode *Node, length int) {
	buf := make([]byte, length)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("[LogForwardHandle] Error reading data:", err)
		return
	}

	logForward := protocol.NewLogForward()
	err = logForward.UNMarshal(buf[:n])
	if err != nil {
		log.Println("[LogForwardHandle] Error unmarshaling data:", err)
		return
	}

	// 判断当前节点是否是Leader，如果是Leader则处理日志条目
	if myNode.Status == LEADER {
		if logForward.Entries != nil {
			for _, entry := range logForward.Entries {
				// 为确保一致性, 只有当前的Term的命令有效
				if entry.Term == myNode.CurrentTerm {
					myNode.Log.AddLog(myNode.CurrentTerm, entry.Command)
				}
			}
			//log.Printf("[LogForwardHandle] Received log entries: %v from node %s, appended to log", logForward.Entries, logForward.NodeID)

		}
	} else {
		log.Printf("[LogForwardHandle] Received log entries but current node is not a Leader, Current Leader: %v\n", myNode.ALLNode.GetRPCListenAddress(myNode.CurrentLeader))
	}
}
