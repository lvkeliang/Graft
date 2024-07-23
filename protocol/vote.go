package protocol

type RequestVote struct {
	term         int64  //leader的term
	candidateIP  string //candidate的IP
	lastLogIndex int64  //candidate最新的log的index
	lastLogTerm  int64  //candidate最新的log的term
}

type RequestVoteResult struct {
	term        int64 //现在的term,用于candidate更新其自己的term
	voteGranted bool  //candidate接受了投票则为true
}
