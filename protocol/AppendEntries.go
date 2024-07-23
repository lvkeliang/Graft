package protocol

import (
	"encoding/json"
	"github.com/lvkeliang/Graft/LogEntry"
)

// AppendEntries 心跳以及日志同步, 首字节标识为1
type AppendEntries struct {
	Term         int64  //Leader的term
	LeaderIP     string //leader的IP
	PrevLogIndex int64  //上一个log的index

	PrevLogTerm int64            ////上一个log的term
	Entries     LogEntry.LogItem //leader的log条目，可一次发送多个以提高效率

	LeaderCommit LogEntry.CommitIndex //Leader的commitIndex
}

func NewAppendEntries() *AppendEntries {
	return &AppendEntries{}
}

// Marshal AppendEntries 首字节标识为1
func (ae *AppendEntries) Marshal() ([]byte, error) {
	marshalAE, err := json.Marshal(ae)
	modifiedData := make([]byte, len(marshalAE)+1)

	// Set the first byte to indicate AppendEntries message type (00000001)
	modifiedData[0] = 1

	// Copy the rest of the data from 'marshalAE' to 'modifiedData'
	copy(modifiedData[1:], marshalAE)
	return modifiedData, err
}

// UNMarshal strae需要先去掉首字节
func (ae *AppendEntries) UNMarshal(strae []byte) error {
	return json.Unmarshal(strae, ae)
}

// AppendEntriesResult 首字节标识为2
type AppendEntriesResult struct {
	Term    int64 //现在的term,用于leader更新其自己的term
	Success bool  //如果follower包含与prevLogIndex和prevLogTerm匹配的entry则为True
}

func NewAppendEntriesResult() *AppendEntriesResult {
	return &AppendEntriesResult{}
}

// Marshal 首字节标识为2
func (aer *AppendEntriesResult) Marshal() ([]byte, error) {
	marshalAE, err := json.Marshal(aer)
	modifiedData := make([]byte, len(marshalAE)+1)

	// Set the first byte to indicate AppendEntries message type (00000001)
	modifiedData[0] = 2

	// Copy the rest of the data from 'marshalAE' to 'modifiedData'
	copy(modifiedData[1:], marshalAE)
	return modifiedData, err
}

// UNMarshal straer需要先去掉首字节
func (aer *AppendEntriesResult) UNMarshal(straer []byte) error {
	return json.Unmarshal(straer, aer)
}
