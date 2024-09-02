package main

import (
	"fmt"
	"github.com/lvkeliang/Graft"
	"github.com/lvkeliang/Graft/stateMachine"
	"log"
	"net/http"
	"time"
)

var myNode *Graft.Node

const RPCPort = "253"
const HTTPPort = "1253"

func main() {
	// 创建节点实例
	myNode = Graft.NewNode(RPCPort, "node_state"+RPCPort+".json", "node"+RPCPort+"log.gob", stateMachine.NewSimpleStateMachine(RPCPort+".txt"))

	// 启动RPC服务器
	err := myNode.StartServer()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// 初始化节点
	myNode.Connect([]string{"localhost:256"})

	// 启动HTTP服务器
	go startHTTPServer(":" + HTTPPort)

	//go termWatcher()

	var sleep chan bool
	<-sleep
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
				log.Printf("[TermWatcher] %v with nodesPool: %v | leaderNow: %v\n", watcher, myNode.ALLNode.Conns, myNode.CurrentLeader)
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

	myNode.AddLogEntry(command)
	fmt.Fprintf(w, "Log added: %s", command)
}
