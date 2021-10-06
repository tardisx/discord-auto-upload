package log

import (
	"log"
	"time"
)

type LogEntryType string

type LogEntry struct {
	Timestamp time.Time    `json:"ts"`
	Type      LogEntryType `json:"type"`
	Entry     string       `json:"log"`
}

const (
	LogTypeInfo  = "info"
	LogTypeError = "error"
	LogTypeDebug = "debug"
)

var LogEntries []LogEntry
var logInput chan LogEntry

func init() {
	// wait for log entries
	logInput = make(chan LogEntry)
	go func() {
		for {
			aLog := <-logInput
			LogEntries = append(LogEntries, aLog)
			for len(LogEntries) > 100 {
				LogEntries = LogEntries[1:]
			}
		}
	}()
}

func SendLog(entry string, entryType LogEntryType) {

	logInput <- LogEntry{
		Timestamp: time.Now(),
		Entry:     entry,
		Type:      entryType,
	}
	log.Printf(entry)
}
