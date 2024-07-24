package LogEntry

// LogItem[term][idx]->command
type LogItem [][]string

type lastApplied struct {
	Term int64
	Idx  int64
}

// IP:prevLogTerm and predLogIndex
type prevLog map[string]struct {
	Term int64
	Idx  int64
}

type CommitIndex struct {
	Term int64
	Idx  int64
}

type LastLog struct {
	Term int64
	Idx  int64
}

type LogEntry struct {
	LogItem     LogItem
	LastApplied lastApplied
	CommitIndex CommitIndex
	LastLog     LastLog
	PrevLog     prevLog
}

func NewLogEntry() *LogEntry {
	return &LogEntry{
		LogItem: LogItem{},
		LastApplied: lastApplied{
			Term: 0,
			Idx:  0,
		},
		CommitIndex: CommitIndex{
			Term: 0,
			Idx:  0,
		},
		PrevLog: prevLog{},
	}
}
