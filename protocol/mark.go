package protocol

const (
	HandShakeMark byte = iota
	HandShakeResultMark
	RequestVoteMark
	RequestVoteResultMark
	AppendEntriesMark
	AppendEntriesResultMark
	LogForwardMark
)
