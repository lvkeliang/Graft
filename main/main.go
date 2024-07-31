package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/lvkeliang/Graft/LogEntry"
	"github.com/lvkeliang/Graft/RPC"
	"github.com/lvkeliang/Graft/matchIndex"
	"github.com/lvkeliang/Graft/nextIndex"
	"github.com/lvkeliang/Graft/node"
	"log"
	"net"
	"time"
)

var myNode *node.Node
var matchIdx *matchIndex.MatchIndex
var nextIdx *nextIndex.NextIndex
var logEnt *LogEntry.LogEntry

func main() {

	go func() {
		err := StartServer(":256")
		if err != nil {
			return
		}
	}()

	go inputNode()

	go termWatcher()

	Init([]string{"localhost:255", "localhost:254", "localhost:253"})
	go RPC.StartElection(context.Background(), myNode, logEnt)
	RPC.StartAppendEntries(context.Background(), myNode, logEnt)

}

func Init(address []string) {
	myNode = node.NewNode()
	matchIdx = matchIndex.NewMatchIndex()
	nextIdx = nextIndex.NewNextIndex()
	logEnt = LogEntry.NewLogEntry()

	for _, addr := range address {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Printf("[connect] connetct to node %v failed\n", addr)
			continue
		}

		myNode.ALLNode.Add(conn)
		go RPC.Handle(conn, myNode, logEnt)
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

		go RPC.Handle(conn, myNode, logEnt)
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
		go RPC.Handle(conn, myNode, logEnt)
	}
}
