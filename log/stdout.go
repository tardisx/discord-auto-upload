package log

import (
	"log"
)

type StdoutLogger struct {
}

func (m StdoutLogger) WriteEntry(l LogEntry) {
	log.Printf("%-6s %s", l.Type, l.Entry)
}
