package main

import (
	"context"
	"errors"
	"github.com/lvkeliang/Graft/LogEntry"
	"github.com/lvkeliang/Graft/RPC"
	"github.com/lvkeliang/Graft/matchIndex"
	"github.com/lvkeliang/Graft/nextIndex"
	"github.com/lvkeliang/Graft/node"
	"log"
	"net"
)

var myNode *node.Node
var matchIdx *matchIndex.MatchIndex
var nextIdx *nextIndex.NextIndex
var logEnt *LogEntry.LogEntry

func main() {
	Start([]string{"localhost"}, ":256")
	go func() {
		err := StartServer(":253")
		if err != nil {
			return
		}
	}()
	RPC.StartAppendEntries(context.Background(), myNode, logEnt)

}

func Start(address []string, port string) {
	myNode = node.NewNode()
	matchIdx = matchIndex.NewMatchIndex()
	nextIdx = nextIndex.NewNextIndex()
	logEnt = LogEntry.NewLogEntry()

	for _, addr := range address {
		conn, err := net.Dial("tcp", addr+port)
		if err != nil {
			log.Printf("[connect] connetct to node %v failed\n", addr+port)
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
		ip := conn.RemoteAddr().String()
		myNode.ALLNode.Add(conn)
		if err != nil {
			log.Printf("[server] accept dial failed\n")
			return errors.New("accept dial failed")
		}

		log.Printf("[server] serving to %v\n", ip)
		log.Printf("[server] nodes now: %v\n", myNode.ALLNode.Conns)

		go RPC.Handle(conn, myNode)
	}
}
