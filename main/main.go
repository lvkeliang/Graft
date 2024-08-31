package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/lvkeliang/Graft/RPC"
	"github.com/lvkeliang/Graft/node"
	"github.com/lvkeliang/Graft/stateMachine"
	"log"
	"net"
	"net/http"
	"time"
)

var myNode *node.Node

const RPCPort = "256"
const HTTPPort = "1256"

func main() {

	go func() {
		err := StartServer(":" + RPCPort)
		if err != nil {
			return
		}
	}()

	//Init([]string{"localhost:254", "localhost:256", "localhost:255"})
	Init([]string{"localhost:255"})

	// 启动HTTP服务器
	go startHTTPServer(":" + HTTPPort)

	go termWatcher()
	go RPC.StartElection(context.Background(), myNode)

	//for i := 1; i < 10; i++ {
	//	myNode.Log.AddLog(myNode.CurrentTerm, "Set x = "+fmt.Sprintf("%d", i))
	//}

	RPC.StartAppendEntries(context.Background(), myNode)

}

func Init(address []string) {
	myNode = node.NewNode(RPCPort, "node_state"+RPCPort+".json", "node"+RPCPort+"log.gob", stateMachine.NewSimpleStateMachine(RPCPort+".txt"))

	for _, addr := range address {
		conn := myNode.Connect(addr)
		if conn != nil {
			RPC.StartHandShake("send", conn, myNode)
			go RPC.Handle(conn, myNode)
		}

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
		if err != nil {
			log.Printf("[server] accept dial failed\n")
			return errors.New("accept dial failed")
		}

		// log.Printf("[server] serving to %v\n", ip)
		// log.Printf("[server] nodes now: %v\n", myNode.ALLNode.Conns)
		RPC.StartHandShake("listen", conn, myNode)
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
				log.Printf("[TermWatcher] %v with nodesPool: %v\n", watcher, myNode.ALLNode.Conns)
			}
		}
	}
}

//func inputNode() {
//	for {
//		addr := ""
//		fmt.Scan(&addr)
//
//		conn, err := net.Dial("tcp", addr)
//		if err != nil {
//			log.Printf("[connect] connetct to node %v failed\n", addr)
//			continue
//		}
//
//		myNode.ALLNode.Add(conn)
//		go RPC.Handle("send", conn, myNode)
//	}
//}

// startHTTPServer 启动HTTP服务器并处理日志添加请求
func startHTTPServer(port string) {
	http.HandleFunc("/log", logHandler)
	log.Fatal(http.ListenAndServe(port, nil)) // 在8080端口启动HTTP服务器
}

// logHandler 处理来自HTTP请求的日志添加
func logHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	command := r.FormValue("command")
	if command == "" {
		http.Error(w, "Command is required", http.StatusBadRequest)
		return
	}

	myNode.Log.AddLog(myNode.CurrentTerm, command)
	fmt.Fprintf(w, "Log added: %s", command)
}
