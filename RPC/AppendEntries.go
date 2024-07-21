package RPC

import (
	"context"
	"github.com/lvkeliang/Graft/node"
	"net"
	"time"
)

func StartAppendEntries(ctx context.Context, pool node.NodesPool) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pool.Lock()
			for _, conn := range pool.Conns {
				conn.Write([]byte{})
				//TODO: 创建一个AppendEntries协议，发送heartbeat以及log
			}
			pool.Unlock()
		}
	}
}

func HeartbeatHandle(conn net.Conn) {
	//TODO: 按照heartbeat协议接收并处理
}
