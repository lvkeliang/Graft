package LogEntry

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/edsrzf/mmap-go"
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
	mu       sync.Mutex
	Entries  []LogEntry
	mmapData mmap.MMap
	filePath string
}

// NewLog initializes a new Log.
func NewLog(filePath string) (*Log, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if stat.Size() == 0 {
		err = file.Truncate(1024 * 1024) // 1MB initial size
		if err != nil {
			return nil, err
		}
	}

	mmapData, err := mmap.Map(file, mmap.RDWR, 0)
	if err != nil {
		return nil, err
	}

	log := &Log{
		Entries:  make([]LogEntry, 0),
		mmapData: mmapData,
		filePath: filePath,
	}

	err = log.load()
	if err != nil {
		return nil, err
	}

	return log, nil
}

func (l *Log) load() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.mmapData) == 0 {
		return nil
	}

	buffer := bytes.NewReader(l.mmapData)
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(&l.Entries)
	if err != nil {
		if err.Error() == "EOF" {
			fmt.Println("Reached end of file, no entries to load.")
			return nil
		}
		fmt.Printf("Error decoding log entries: %v\n", err)
		return err
	}

	return nil
}

func (l *Log) Append(entry LogEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Entries = append(l.Entries, entry)
	return l.persist()
}

func (l *Log) persist() error {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(l.Entries)
	if err != nil {
		return err
	}

	if len(buffer.Bytes()) > len(l.mmapData) {
		// Unmap, resize file, remap
		err = l.resizeFile(len(buffer.Bytes()))
		if err != nil {
			return err
		}
	}

	copy(l.mmapData, buffer.Bytes())
	return nil
}

func (l *Log) resizeFile(newSize int) error {
	l.mmapData.Unmap()
	file, err := os.OpenFile(l.filePath, os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	err = file.Truncate(int64(newSize))
	if err != nil {
		return err
	}

	l.mmapData, err = mmap.Map(file, mmap.RDWR, 0)
	if err != nil {
		return err
	}

	return nil
}

func (l *Log) Get(index int64) (*LogEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if index < 0 || index >= int64(len(l.Entries)) {
		return nil, fmt.Errorf("index out of range")
	}

	return &l.Entries[index], nil
}

func (l *Log) LastIndex() int64 {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.Entries) == 0 {
		return -1
	}

	return l.Entries[len(l.Entries)-1].Index
}

func (l *Log) GetLastEntry() *LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.Entries) == 0 {
		return nil
	}

	return &l.Entries[len(l.Entries)-1]
}

func (l *Log) Truncate(index int64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if index < 0 || index >= int64(len(l.Entries)) {
		return fmt.Errorf("index out of range")
	}

	l.Entries = l.Entries[:index+1]
	return l.persist()
}

func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.mmapData.Unmap()
}

func main() {
	log, err := NewLog("raft_log")
	if err != nil {
		fmt.Println("Error initializing log:", err)
		return
	}
	defer log.Close()

	entry := LogEntry{Term: 1, Index: 1, Command: "set x=1"}
	err = log.Append(entry)
	if err != nil {
		fmt.Println("Error appending log entry:", err)
	}

	retrievedEntry, err := log.Get(1)
	if err != nil {
		fmt.Println("Error getting log entry:", err)
	} else {
		fmt.Printf("Retrieved log entry: %+v\n", retrievedEntry)
	}

	fmt.Println("Last index:", log.LastIndex())
}
