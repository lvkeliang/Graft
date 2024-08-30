package RPC

import (
	"github.com/lvkeliang/Graft/node"
	"github.com/lvkeliang/Graft/protocol"
	"log"
	"net"
)

func SendHandshake(RPCListenPort string, conn net.Conn) error {
	// 发送当前节点的监听端口信息到服务器
	_, err := conn.Write([]byte(RPCListenPort))
	if err != nil {
		log.Println("[SendHandshake] Error sending node info:", err)
		return err
	}

	handshake := protocol.NewHandshake()
	handshake = &protocol.Handshake{
		RPCListenPort: RPCListenPort,
	}

	marshalHS, err := handshake.Marshal()
	// fmt.Println(string(marshalAE))
	if err != nil {
		log.Println("[SendHandshake] appendEntries.Marshal failed", err)
		return err
	}

	_, err = conn.Write(marshalHS)
	if err != nil {
		log.Println("[SendHandshake] send marshalAE failed: ", err)
		return err
	}

	return nil
}

func HandshakeHandle(conn net.Conn, myNode *node.Node, length int) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("[HandshakeHandle] Error reading node info:", err)
		return
	}

	nodeInfo := string(buf[:n])
	log.Println("[HandshakeHandle] Received node info:", nodeInfo)
}

func HandshakeHandleResult(conn net.Conn, myNode *node.Node, length int) {

}
