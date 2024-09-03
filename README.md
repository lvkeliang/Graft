### English Version

[切换到中文版](#中文版)

# Graft: A Simple Raft Implementation in Go

Graft is a simple, yet effective implementation of the Raft consensus protocol, written in Go. It allows a distributed system to maintain a consistent log across multiple nodes, ensuring that each node agrees on the state of the system.

## Features

- **Leader Election**: Handles the process of electing a leader among the nodes in the cluster.
- **Log Replication**: Ensures that the leader node replicates log entries across the follower nodes.
- **Persistent State**: Maintains persistent state, ensuring consistency even after node restarts.
- **State Machine Integration**: Provides an interface for state machines, allowing you to apply log entries to a state machine.
- **Cluster Connectivity**: New nodes can join the cluster by connecting to any existing node, automatically connecting to all nodes in the cluster.
- **Log Forwarding**: Non-leader nodes can add log entries, which are automatically forwarded to the leader for processing.
- **Fault Tolerance**: Tolerates node failures and ensures the cluster can still reach consensus.

## Code Highlights

1. **Interface Provided**:
   - Graft provides the `StateMachine` interface, enabling users to easily implement their own state machine logic. By implementing the `ToApply` method, users can customize how log entries are applied and how the state is persisted.

2. **State Machine Integration**:
   - Graft provides a simple state machine implementation, `SimpleStateMachine`, demonstrating how to persist and manage state.

3. **Automatic Cluster Connectivity**:
   - New nodes can join the cluster by connecting to any existing node. Graft’s `myNode.Connect` method automatically connects the new node to all other nodes in the cluster, simplifying cluster setup.

4. **Log Forwarding**:
   - Graft’s `myNode.AddLogEntry` method allows non-leader nodes to add log entries, which are then automatically forwarded to the leader node for execution, ensuring that the log is consistently replicated.

## Installation

To use Graft in your project, you can clone the repository and install the necessary dependencies:

```bash
git clone https://github.com/lvkeliang/Graft.git
cd Graft
go mod tidy
```

Alternatively, you can use `go get` to install Graft:

```bash
go get github.com/lvkeliang/Graft
```

## Usage

### Creating a New Node

You can create a new Raft node by calling the `NewNode` function. Here’s an example using a simple state machine:

```go
import (
    "github.com/lvkeliang/Graft"
    "grafttest/statemachine"
)

func main() {
    stateMachine := statemachine.NewSimpleStateMachine("state_file.txt")
    node := Graft.NewNode(":8080", "state_file.json", "log_file.gob", stateMachine)

    if node != nil {
        node.StartServer()
    }
}
```

### Connecting Nodes

Graft makes it easy to add new nodes to the cluster. Simply connect a new node to any existing node in the cluster, and it will automatically connect to all other nodes:

```go
node.Connect([]string{"127.0.0.1:254"})
```

### Adding Log Entries

Log entries can be added directly using the `node.AddLogEntry` method. Non-leader nodes will automatically forward the log entries to the current leader for processing:

```go
import (
    "fmt"
    "log"
    "net/http"
    "github.com/lvkeliang/Graft"
)

var myNode *Graft.Node

func main() {
    myNode = Graft.NewNode(":255", "node_state255.json", "node255log.gob", statemachine.NewSimpleStateMachine("255.txt"))

    if err := myNode.StartServer(); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }

    myNode.Connect([]string{"127.0.0.1:254"})

    startHTTPServer(":1255")
}

func startHTTPServer(port string) {
    http.HandleFunc("/log", logHandler)
    log.Fatal(http.ListenAndServe(port, nil))
}

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
```

### State Machine Interface

Graft provides a `StateMachine` interface that you can implement to define how log entries are applied:

```go
package stateMachine

type StateMachine interface {
    ToApply(command interface{}) interface{}
}
```

Here’s an example implementation:

```go
package statemachine

import (
    "fmt"
    "io/ioutil"
)

type SimpleStateMachine struct {
    filePath string
    state    interface{}
}

func NewSimpleStateMachine(filePath string) *SimpleStateMachine {
    return &SimpleStateMachine{
        filePath: filePath,
    }
}

func (sm *SimpleStateMachine) ToApply(command interface{}) interface{} {
    sm.state = command
    err := ioutil.WriteFile(sm.filePath, []byte(fmt.Sprintf("%v", sm.state)), 0644)
    if err != nil {
        fmt.Printf("[SimpleStateMachine] failed to write state to file: %v\n", err)
        return nil
    }
    fmt.Printf("[SimpleStateMachine] State written to file: %v\n", sm.state)
    return sm.state
}

func (sm *SimpleStateMachine) GetState() interface{} {
    data, err := ioutil.ReadFile(sm.filePath)
    if err != nil {
        fmt.Printf("[SimpleStateMachine] failed to read state from file: %v\n", err)
        return nil
    }
    sm.state = string(data)
    return sm.state
}
```

### Example Cluster Setup

Here’s a simple example of setting up a Graft cluster:

```go
package main

import (
    "log"
    "github.com/lvkeliang/Graft"
    "grafttest/statemachine"
)

func main() {
    myNode := Graft.NewNode(":255", "node_state255.json", "node255log.gob", statemachine.NewSimpleStateMachine("255.txt"))

    // Start the RPC server
    if err := myNode.StartServer(); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }

    // Connect to other nodes in the cluster
    myNode.Connect([]string{"127.0.0.1:254"})

    // Start the HTTP server to handle log entries
    startHTTPServer(":1255")
}
```

---

### 中文版

[Switch to English Version](#English-Version)

# Graft: 一个用Go实现的简单Raft协议

Graft是一个用Go语言编写的简单而有效的Raft一致性协议实现。它允许分布式系统在多个节点之间维护一致的日志，确保每个节点对系统的状态达成一致。

## 功能

- **领导者选举**：处理集群中节点之间的领导者选举过程。
- **日志复制**：确保领导者节点将日志条目复制到所有跟随者节点。
- **持久化状态**：维护持久化状态，即使节点重启也能确保一致性。
- **状态机集成**：提供了状态机接口，允许您将日志条目应用到状态机。
- **集群连接**：新节点可以通过连接集群中的任何一个现有节点来加入集群，并自动连接到集群中的所有节点。
- **日志转发**：非领导者节点可以添加日志条目，日志将自动转发到领导者进行处理。
- **容错性**：能够容忍节点故障，确保集群仍能达成一致性。

## 代码特色

1. **提供接口**：
   - Graft提供了`StateMachine`接口，使用户能够轻松实现自己的状态机逻辑。通过实现`ToApply`方法，用户可以自定义如何应用日志条目以及如何持久化状态。

2. **状态机集成**：
   - Graft提供了一个简单的状态机实现`SimpleStateMachine`，展示了如何持久化和管理状态。

3. **自动集群连接**：
   - 新节点可以通过连接集群中的任意一个现有节点来加入集群。Graft的`myNode.Connect`方法会自动将新节点连接到集群中的所有其他节点，简化了集群的设置过程。

4. **日志转发**：
   - Graft的`myNode.AddLogEntry`方法允许非领导者节点添加日志条目，这些日志条目会自动转发到领导者节点执行，确保日志的一致性复制。

## 安装

要在项目中使用Graft，您可以克隆仓库并安装必要的依赖：

```bash
git clone https://github.com/lvkeliang/Graft.git
cd Graft
go mod tidy
```

或者，您可以使用`go get`来安装Graft：

```bash
go get github.com/lvkeliang/Graft
```

## 使用

### 创建一个新节点

您可以

通过调用`NewNode`函数来创建一个新的Raft节点。以下是一个使用简单状态机的示例：

```go
import (
    "github.com/lvkeliang/Graft"
    "grafttest/statemachine"
)

func main() {
    stateMachine := statemachine.NewSimpleStateMachine("state_file.txt")
    node := Graft.NewNode(":8080", "state_file.json", "log_file.gob", stateMachine)

    if node != nil {
        node.StartServer()
    }
}
```

### 连接节点

Graft让您可以轻松地将新节点添加到集群中。只需将新节点连接到集群中的任意一个现有节点，它就会自动连接到所有其他节点：

```go
node.Connect([]string{"127.0.0.1:254"})
```

### 添加日志条目

日志条目可以直接使用`node.AddLogEntry`方法添加。非领导者节点会自动将日志条目转发给当前的领导者进行处理：

```go
import (
    "fmt"
    "log"
    "net/http"
    "github.com/lvkeliang/Graft"
)

var myNode *Graft.Node

func main() {
    myNode = Graft.NewNode(":255", "node_state255.json", "node255log.gob", statemachine.NewSimpleStateMachine("255.txt"))

    if err := myNode.StartServer(); err != nil {
        log.Fatalf("服务器启动失败: %v", err)
    }

    myNode.Connect([]string{"127.0.0.1:254"})

    startHTTPServer(":1255")
}

func startHTTPServer(port string) {
    http.HandleFunc("/log", logHandler)
    log.Fatal(http.ListenAndServe(port, nil))
}

func logHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "只允许POST方法", http.StatusMethodNotAllowed)
        return
    }

    command := r.FormValue("command")
    if command == "" {
        http.Error(w, "命令是必需的", http.StatusBadRequest)
        return
    }

    myNode.AddLogEntry(command)
    fmt.Fprintf(w, "日志已添加: %s", command)
}
```

### 状态机接口

Graft提供了一个`StateMachine`接口，您可以实现该接口来定义如何应用日志条目：

```go
package stateMachine

type StateMachine interface {
    ToApply(command interface{}) interface{}
}
```

以下是一个示例实现：

```go
package statemachine

import (
    "fmt"
    "io/ioutil"
)

type SimpleStateMachine struct {
    filePath string
    state    interface{}
}

func NewSimpleStateMachine(filePath string) *SimpleStateMachine {
    return &SimpleStateMachine{
        filePath: filePath,
    }
}

func (sm *SimpleStateMachine) ToApply(command interface{}) interface{} {
    sm.state = command
    err := ioutil.WriteFile(sm.filePath, []byte(fmt.Sprintf("%v", sm.state)), 0644)
    if err != nil {
        fmt.Printf("[SimpleStateMachine] 将状态写入文件失败: %v\n", err)
        return nil
    }
    fmt.Printf("[SimpleStateMachine] 状态已写入文件: %v\n", sm.state)
    return sm.state
}

func (sm *SimpleStateMachine) GetState() interface{} {
    data, err := ioutil.ReadFile(sm.filePath)
    if err != nil {
        fmt.Printf("[SimpleStateMachine] 从文件读取状态失败: %v\n", err)
        return nil
    }
    sm.state = string(data)
    return sm.state
}
```

### 集群设置示例

以下是一个设置Graft集群的简单示例：

```go
package main

import (
    "log"
    "github.com/lvkeliang/Graft"
    "grafttest/statemachine"
)

func main() {
    myNode := Graft.NewNode(":255", "node_state255.json", "node255log.gob", statemachine.NewSimpleStateMachine("255.txt"))

    // 启动RPC服务器
    if err := myNode.StartServer(); err != nil {
        log.Fatalf("服务器启动失败: %v", err)
    }

    // 连接到集群中的其他节点
    myNode.Connect([]string{"127.0.0.1:254"})

    // 启动HTTP服务器以处理日志条目
    startHTTPServer(":1255")
}
```