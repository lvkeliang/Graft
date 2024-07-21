package LogEntry

type item struct {
	term    int64
	idx     int64
	command string
}

type LogEntry struct {
	logItem     []item
	LastApplied int64
	CommitIndex int64
}

func NewLogEntry() *LogEntry {
	return &LogEntry{
		logItem:     []item{},
		LastApplied: 0,
		CommitIndex: 0,
	}
}
