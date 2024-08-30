package RPC

import (
	"fmt"
	"github.com/lvkeliang/Graft/node"
	"github.com/lvkeliang/Graft/protocol"
	"log"
	"net"
)

func Handle(conn net.Conn, myNode *node.Node) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("[Handle] conn close error")
			return
		}
	}(conn)

	for {
		//buf := make([]byte, 1024*5)
		//n, _ := conn.Read(buf)
		//log.Println("Unknown type message:", string(buf[:n]))

		// Read the first byte to determine the message type
		msgType := make([]byte, 1)
		_, err := conn.Read(msgType)
		if err != nil {
			log.Println("[Handle] Error reading message type:", err)
			myNode.RemoveNode(conn)
			return
		}

		lengthByte := make([]byte, 2)
		_, err = conn.Read(lengthByte)
		if err != nil {
			log.Println("[Handle] Error reading message lengthByte:", err)
			myNode.RemoveNode(conn)
			return
		}

		length := protocol.BytesToInt(lengthByte)

		switch msgType[0] {
		case protocol.AppendEntriesMark:
			// Handle AppendEntries message
			AppendEntriesHandle(conn, myNode, length)
			fmt.Println("消息类型：1")
		case protocol.AppendEntriesResultMark:
			AppendEntriesResultHandle(conn, myNode, length)
			fmt.Println("消息类型：2")
		case protocol.RequestVoteMark:
			RequestVoteHandle(conn, myNode, length)
			fmt.Println("消息类型：3")
		case protocol.RequestVoteResultMark:
			RequestVoteResultHandle(conn, myNode, length)
			fmt.Println("消息类型：4")
		default:
			log.Println("[Handle] Unknown message type:", msgType[0])
			//buf := make([]byte, length)
			//n, _ := conn.Read(buf)
			//log.Println("Unknown type message:", string(buf[:n]))
		}
	}

}
