package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/lvkeliang/Graft/RPC"
	"github.com/lvkeliang/Graft/node"
	"log"
	"net"
	"time"
)

var myNode *node.Node

func main() {

	go func() {
		err := StartServer(":253")
		if err != nil {
			return
		}
	}()

	Init([]string{"localhost:255", "localhost:256", "localhost:254"})

	go inputNode()

	go termWatcher()
	go RPC.StartElection(context.Background(), myNode)
	RPC.StartAppendEntries(context.Background(), myNode)

}

func Init(address []string) {
	myNode = node.NewNode()

	for _, addr := range address {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Printf("[connect] connetct to node %v failed\n", addr)
			continue
		}

		myNode.ALLNode.Add(conn)
		go RPC.Handle(conn, myNode)
	}
}

func StartServer(port string) error {
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("[server] start serve on port %v failed:%v\n", port, err.Error())
		return errors.New("start serve failed")
	}

	log.Printf("[server] serving on port %v\n", port)

	for {
		conn, err := ln.Accept()
		// ip := conn.RemoteAddr().String()
		myNode.ALLNode.Add(conn)
		if err != nil {
			log.Printf("[server] accept dial failed\n")
			return errors.New("accept dial failed")
		}

		// log.Printf("[server] serving to %v\n", ip)
		// log.Printf("[server] nodes now: %v\n", myNode.ALLNode.Conns)

		go RPC.Handle(conn, myNode)
	}
}

func termWatcher() {
	ticker := time.NewTicker(1 * time.Second)
	watcher := myNode.CurrentTerm
	log.Printf("[TermWatcher] %v\n", watcher)
	for {
		select {
		case <-ticker.C:
			if myNode.CurrentTerm != watcher {
				watcher = myNode.CurrentTerm
				log.Printf("[TermWatcher] %v\n", watcher)
			}
		}
	}
}

func inputNode() {
	for {
		addr := ""
		fmt.Scan(&addr)

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Printf("[connect] connetct to node %v failed\n", addr)
			continue
		}

		myNode.ALLNode.Add(conn)
		go RPC.Handle(conn, myNode)
	}
}
