package log

import (
	"fmt"
	"time"
)

type Logger interface {
	WriteEntry(l LogEntry)
}

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

var loggers []Logger
var logInput chan LogEntry
var Memory *MemoryLogger

func init() {

	// create some loggers
	Memory = &MemoryLogger{maxsize: 100}
	stdout := &StdoutLogger{}

	loggers = []Logger{Memory, stdout}

	// wait for log entries
	logInput = make(chan LogEntry)
	go func() {
		for {
			aLog := <-logInput
			for _, l := range loggers {
				l.WriteEntry(aLog)
			}
		}
	}()
}

func Debug(entry string) {
	logInput <- LogEntry{
		Timestamp: time.Now(),
		Entry:     entry,
		Type:      LogTypeDebug,
	}
}
func Debugf(entry string, args ...interface{}) {
	logInput <- LogEntry{
		Timestamp: time.Now(),
		Entry:     fmt.Sprintf(entry, args...),
		Type:      LogTypeDebug,
	}
}

func Info(entry string) {
	logInput <- LogEntry{
		Timestamp: time.Now(),
		Entry:     entry,
		Type:      LogTypeInfo,
	}
}

func Infof(entry string, args ...interface{}) {
	logInput <- LogEntry{
		Timestamp: time.Now(),
		Entry:     fmt.Sprintf(entry, args...),
		Type:      LogTypeInfo,
	}
}

func Error(entry string) {
	logInput <- LogEntry{
		Timestamp: time.Now(),
		Entry:     entry,
		Type:      LogTypeError,
	}
}

func Errorf(entry string, args ...interface{}) {
	logInput <- LogEntry{
		Timestamp: time.Now(),
		Entry:     fmt.Sprintf(entry, args...),
		Type:      LogTypeError,
	}
}
