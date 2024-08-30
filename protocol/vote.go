package protocol

import "encoding/json"

type RequestVote struct {
	Term         int64  //leader的term
	CandidateIP  string //candidate的IP
	LastLogIndex int64  //candidate最新的log的index
	LastLogTerm  int64  //candidate最新的log的term
}

func NewRequestVote() *RequestVote {
	return &RequestVote{}
}

// Marshal AppendEntries 首字节标识为3
func (rv *RequestVote) Marshal() ([]byte, error) {
	marshalRV, err := json.Marshal(rv)
	modifiedData := make([]byte, len(marshalRV)+3)

	// Set the first byte to indicate AppendEntries message type (00000001)
	modifiedData[0] = RequestVoteMark

	copy(modifiedData[1:3], IntToBytes(len(marshalRV)))

	// Copy the rest of the data from 'marshalAE' to 'modifiedData'
	copy(modifiedData[3:], marshalRV)
	return modifiedData, err
}

// UNMarshal strae需要先去掉首字节
func (rv *RequestVote) UNMarshal(strrv []byte) error {
	return json.Unmarshal(strrv, rv)
}

type RequestVoteResult struct {
	Term        int64 //现在的term,用于candidate更新其自己的term
	VoteGranted bool  //candidate接受了投票则为true
}

func NewRequestVoteResult() *RequestVoteResult {
	return &RequestVoteResult{}
}

// Marshal AppendEntries 首字节标识为4
func (rvr *RequestVoteResult) Marshal() ([]byte, error) {
	marshalRVR, err := json.Marshal(rvr)
	modifiedData := make([]byte, len(marshalRVR)+3)

	// Set the first byte to indicate AppendEntries message type (00000001)
	modifiedData[0] = RequestVoteResultMark

	copy(modifiedData[1:3], IntToBytes(len(marshalRVR)))

	// Copy the rest of the data from 'marshalAE' to 'modifiedData'
	copy(modifiedData[3:], marshalRVR)
	return modifiedData, err
}

// UNMarshal strae需要先去掉首字节
func (rvr *RequestVoteResult) UNMarshal(strrvr []byte) error {
	return json.Unmarshal(strrvr, rvr)
}
