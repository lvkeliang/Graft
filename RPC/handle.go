package RPC

import (
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

		// Read the first byte to determine the message type
		msgType := make([]byte, 1)
		_, err := conn.Read(msgType)
		if err != nil {
			log.Println("[Handle] Error reading message type:", err)
			myNode.ALLNode.Remove(conn)
			return
		}

		lengthByte := make([]byte, 2)
		_, err = conn.Read(lengthByte)
		if err != nil {
			log.Println("[Handle] Error reading message lengthByte:", err)
			myNode.ALLNode.Remove(conn)
			return
		}

		length := protocol.BytesToInt(lengthByte)

		switch msgType[0] {
		case 1:
			// Handle AppendEntries message
			AppendEntriesHandle(conn, myNode, length)
		case 2:
			AppendEntriesResultHandle(conn, length)
		case 3:
			RequestVoteHandle(conn, myNode, length)
		case 4:
			RequestVoteResultHandle(conn, myNode, length)
		default:
			log.Println("[Handle] Unknown message type:", msgType[0])
		}
	}

}
