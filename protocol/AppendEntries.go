package protocol

// AppendEntries 心跳以及日志同步
type AppendEntries struct {
	term         int64  //Leader的term
	leaderIP     string //leader的IP
	prevLogIndex int64  //上一个log的index

	prevLogTerm int64    ////上一个log的term
	entries     []string //leader的log条目，可一次发送多个以提高效率

	leaderCommit int64 //Leader的commitIndex
}

type AppendEntriesResult struct {
	term    int64 //现在的term,用于leader更新其自己的term
	success bool  //如果follower包含与prevLogIndex和prevLogTerm匹配的entry则为True
}
