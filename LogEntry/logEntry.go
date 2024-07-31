package LogEntry

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

// LogEntry represents a single entry in the Raft log.
type LogEntry struct {
	Term    int64
	Index   int64
	Command interface{}
}

// Log represents the log in a Raft node.
type Log struct {
	mu      sync.Mutex
	Entries []LogEntry
	file    *os.File
}

// NewLog creates a new log and loads entries from the specified file.
func NewLog(filename string) (*Log, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	log := &Log{
		Entries: make([]LogEntry, 0),
		file:    file,
	}

	// Load existing log entries from file
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if len(data) > 0 {
		err = json.Unmarshal(data, &log.Entries)
		if err != nil {
			return nil, err
		}
	}

	return log, nil
}

// Append appends a new entry to the log.
func (l *Log) Append(entry LogEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Entries = append(l.Entries, entry)
	return l.Save()
}

// GetEntries returns the log entries starting from the specified index.
func (l *Log) GetEntries(fromIndex int) ([]LogEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if fromIndex < 0 || fromIndex >= len(l.Entries) {
		return nil, nil
	}

	return l.Entries[fromIndex:], nil
}

// GetLastEntry returns the last log entry.
func (l *Log) GetLastEntry() *LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.Entries) == 0 {
		return nil
	}

	return &l.Entries[len(l.Entries)-1]
}

// Save persists the log Entries to the file.
func (l *Log) Save() error {
	data, err := json.Marshal(l.Entries)
	if err != nil {
		return err
	}

	if _, err = l.file.Seek(0, 0); err != nil {
		return err
	}

	if err = l.file.Truncate(0); err != nil {
		return err
	}

	_, err = l.file.Write(data)
	return err
}

// Close closes the log file.
func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}
