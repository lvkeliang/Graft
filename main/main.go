package main

import (
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

func Start(address []string, port string) {
	myNode = node.NewNode()
	matchIdx = matchIndex.NewMatchIndex()
	nextIdx = nextIndex.NewNextIndex()
	logEnt = LogEntry.NewLogEntry()

	for _, addr := range address {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Printf("[connect] connetct to node %v failed\n", addr)
		}

		myNode.ALLNode.Add(addr, conn)
	}

	err := StartServer(":256")
	if err != nil {
		return
	}

}

func StartServer(port string) error {
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("[server] start serve on port %v failed\n", port)
		return errors.New("start serve failed")
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[server] accept dial failed\n")
			return errors.New("accept dial failed")
		}
		go RPC.HeartbeatHandle(conn)
	}
}
