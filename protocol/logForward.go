package protocol

import (
	"encoding/json"
	"github.com/lvkeliang/Graft/LogEntry"
)

// LogForward 用于从非Leader节点向Leader转发日志，首字节标识为3
type LogForward struct {
	Term    int64               // 当前节点的Term
	NodeID  string              // 当前节点的ID
	Entries []LogEntry.LogEntry // 要转发的日志条目
}

func NewLogForward() *LogForward {
	return &LogForward{}
}

// Marshal LogForward
func (lf *LogForward) Marshal() ([]byte, error) {
	marshalLF, err := json.Marshal(lf)
	modifiedData := make([]byte, len(marshalLF)+3)

	// Set the first byte to indicate LogForward message type (00000011)
	modifiedData[0] = LogForwardMark

	// 长度字节
	copy(modifiedData[1:3], IntToBytes(len(marshalLF)))

	// Copy the rest of the data from 'marshalLF' to 'modifiedData'
	copy(modifiedData[3:], marshalLF)
	return modifiedData, err
}

// UNMarshal 需要先去掉首字节(1字节)和长度字节(2字节)
func (lf *LogForward) UNMarshal(strae []byte) error {
	return json.Unmarshal(strae, lf)
}
