package Graft

import (
	"github.com/lvkeliang/Graft/protocol"
	"log"
	"net"
)

func readHeader(conn net.Conn, myNode *Node) (msgType byte, length int, err error) {
	msgTypeReader := make([]byte, 1)
	_, err = conn.Read(msgTypeReader)
	if err != nil {
		log.Println("[Handle] Error reading message type:", err)
		myNode.RemoveNode(conn)
		return 0, 0, err
	}

	lengthByte := make([]byte, 2)
	_, err = conn.Read(lengthByte)
	if err != nil {
		log.Println("[Handle] Error reading message lengthByte:", err)
		myNode.RemoveNode(conn)
		return 0, 0, err
	}

	msgType = msgTypeReader[0]
	length = protocol.BytesToInt(lengthByte)
	return msgType, length, nil
}

func Handle(conn net.Conn, myNode *Node) {

	for {
		//buf := make([]byte, 1024*5)
		//n, _ := conn.Read(buf)
		//log.Println("Unknown type message:", string(buf[:n]))

		// Read the first byte to determine the message type
		msgType, length, err := readHeader(conn, myNode)
		if err != nil {
			log.Println("[Handle] ReadHeader Failed")
			return
		}

		switch msgType {
		case protocol.AppendEntriesMark:
			// Handle AppendEntries message
			AppendEntriesHandle(conn, myNode, length)
		case protocol.AppendEntriesResultMark:
			AppendEntriesResultHandle(conn, myNode, length)
		case protocol.RequestVoteMark:
			RequestVoteHandle(conn, myNode, length)
		case protocol.RequestVoteResultMark:
			RequestVoteResultHandle(conn, myNode, length)
		default:
			log.Println("[Handle] Unknown message type:", msgType)
			// 遗弃接下来的消息
			msgTypeReader := make([]byte, 1024)
			_, err = conn.Read(msgTypeReader)
			//buf := make([]byte, length)
			//n, _ := conn.Read(buf)
			//log.Println("Unknown type message:", string(buf[:n]))
		}
	}

}
