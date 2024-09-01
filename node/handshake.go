package node

import (
	"errors"
	"fmt"
	"github.com/lvkeliang/Graft/protocol"
	"log"
	"net"
)

func SendHandshake(myNode *Node, conn net.Conn) error {
	// 发送当前节点的监听端口信息到服务器

	handshake := protocol.NewHandshake()
	handshake = &protocol.Handshake{
		RPCListenPort: myNode.RPCListenPort,
	}

	marshalHS, err := handshake.Marshal()
	if err != nil {
		log.Println("[SendHandshake] appendEntries.Marshal failed", err)
		return err
	}

	_, err = conn.Write(marshalHS)
	if err != nil {
		log.Println("[SendHandshake] send marshalAE failed: ", err)
		return err
	}

	_, port, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return err
	}

	myNode.AddNode(port, conn)

	return nil
}

func HandshakeHandle(conn net.Conn, myNode *Node, length int) error {
	buf := make([]byte, length)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("[HandshakeHandle] Error reading node info:", err)
		return err
	}

	nodeInfo := string(buf[:n])
	log.Println("[HandshakeHandle] Received node info:", nodeInfo)

	// Parse received data into AppendEntriesResult
	handshake := protocol.NewHandshake()
	err = handshake.UNMarshal(buf[:n])
	if err != nil {
		log.Println("[HandshakeHandle] Error unmarshaling data:", err)
		return err
	}

	myNode.AddNode(handshake.RPCListenPort, conn)

	fmt.Println("GetALLRPCListenAddresses: ", myNode.ALLNode.GetALLRPCListenAddresses())
	fmt.Println("Addresses: ", myNode.ALLNode.Conns)

	handshakeResult := protocol.NewHandshakeResult()
	handshakeResult = &protocol.HandshakeResult{
		NodeAddresses: myNode.ALLNode.GetALLRPCListenAddresses(),
		Result:        "success",
	}

	marshalHSR, err := handshakeResult.Marshal()
	if err != nil {
		log.Println("[HandshakeHandle] appendEntries.Marshal failed", err)
		return err
	}

	_, err = conn.Write(marshalHSR)
	if err != nil {
		log.Println("[HandshakeHandle] send marshalAE failed: ", err)
		return err
	}

	return nil
}

func HandshakeHandleResult(conn net.Conn, myNode *Node, length int) error {
	buf := make([]byte, length)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("[HandshakeHandleResult] Error reading node info:", err)
		return err
	}

	nodeInfo := string(buf[:n])
	log.Println("[HandshakeHandleResult] Received node info:", nodeInfo)

	// Parse received data into AppendEntriesResult
	handshakeResult := protocol.NewHandshakeResult()
	err = handshakeResult.UNMarshal(buf[:n])
	if err != nil {
		log.Println("[HandshakeHandleResult] Error unmarshaling data:", err)
		return err
	}

	if handshakeResult.Result != "success" {
		log.Println("[HandshakeHandleResult] handshake failed:", err)
		return err
	}

	for _, addr := range handshakeResult.NodeAddresses {
		//fmt.Println("GETADDRESS : ", addr)

		if addr != "" && addr != "127.0.0.1:"+myNode.RPCListenPort && myNode.ALLNode.Get(addr) == nil {
			fmt.Println("GetALLRPCListenAddresses: ", myNode.ALLNode.GetALLRPCListenAddresses())
			fmt.Println("Addresses: ", myNode.ALLNode.Conns)
			fmt.Println("CONNECT TO : ", addr)

			newConn, err := connectToAddress(addr)
			if err != nil {
				log.Printf("[connect] connetct to node %v failed\n", addr)
				return err
			}

			if newConn != nil {
				StartHandShake("send", newConn, myNode)
				go Handle(newConn, myNode)
			}
		}
	}

	return nil
}

func StartHandShake(handshakeMode string, conn net.Conn, myNode *Node) error {

	if handshakeMode == "send" {
		err := SendHandshake(myNode, conn)
		if err != nil {
			log.Println("[Handle] SendHandshake Failed")
			return err
		}

		// Read the first byte to determine the message type
		msgType, length, err := readHeader(conn, myNode)
		if err != nil || msgType != protocol.HandShakeResultMark {
			log.Println("[Handle] ReadHeader of HandShake Failed")
			return err
		}

		HandshakeHandleResult(conn, myNode, length)
		if err != nil {
			log.Println("[Handle] handshake failed")
			return err
		}

	} else if handshakeMode == "listen" {
		// Read the first byte to determine the message type
		msgType, length, err := readHeader(conn, myNode)
		if err != nil || msgType != protocol.HandShakeMark {
			log.Println("[Handle] ReadHeader of HandShake Failed")
			return err
		}

		err = HandshakeHandle(conn, myNode, length)
		if err != nil {
			log.Println("[Handle] ReadHeader of HandShake Failed")
			return err
		}

	} else {
		log.Println("[Handle] UNKnown handshakeMode")
		return errors.New("[Handle] UNKnown handshakeMode")
	}

	return nil
}
