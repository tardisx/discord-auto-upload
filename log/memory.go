package log

import (
	"sync"
)

type MemoryLogger struct {
	size    int
	entries []LogEntry
	maxsize int
	lock    sync.Mutex
}

func (m *MemoryLogger) WriteEntry(l LogEntry) {
	// xxx needs mutex
	// if m.entries == nil {
	// 	m.entries = make([]LogEntry, 0)
	// }
	m.lock.Lock()
	m.entries = append(m.entries, l)
	if len(m.entries) > m.maxsize {
		m.entries = m.entries[1:]
	}
	m.lock.Unlock()
}

func (m *MemoryLogger) Entries() []LogEntry {
	return m.entries
}
