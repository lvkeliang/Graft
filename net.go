package Graft

import (
	"context"
	"errors"
	"log"
	"net"
)

func connectToAddress(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (node *Node) Connect(addresses []string) {
	for _, addr := range addresses {
		conn, err := connectToAddress(addr)
		if err != nil {
			log.Printf("[connect] connetct to node %v failed\n", addr)
			return
		}

		if conn != nil {
			StartHandShake("send", conn, node)
			go Handle(conn, node)
		}
	}
}

func (node *Node) StartServer() error {
	ln, err := net.Listen("tcp", node.RPCListenPort)
	if err != nil {
		log.Printf("[server] start serve on port %v failed:%v\n", node.RPCListenPort, err.Error())
		return errors.New("start serve failed")
	}

	log.Printf("[server] serving on port %v\n", node.RPCListenPort)

	// 启动选举进程
	go StartElection(context.Background(), node)

	// 启动日志复制进程
	go StartAppendEntries(context.Background(), node)

	// 启动一个新的goroutine来处理连接
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Printf("[server] accept dial failed\n")
				return
			}

			StartHandShake("listen", conn, node)
			go Handle(conn, node)
		}
	}()

	return nil
}
