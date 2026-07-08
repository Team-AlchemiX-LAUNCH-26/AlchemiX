package audit

import (
	"encoding/json"
	"sync"
)

// Logger is a thread-safe audit log for routing decisions.
type Logger struct {
	mu      sync.RWMutex
	records []Record
	maxSize int
}

// NewLogger creates a logger with the given buffer size.
func NewLogger(maxSize int) *Logger {
	if maxSize < 1 {
		maxSize = 1000
	}
	return &Logger{records: make([]Record, 0), maxSize: maxSize}
}

// Log adds a record to the audit log.
func (l *Logger) Log(record Record) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.records = append(l.records, record)
	if len(l.records) > l.maxSize {
		l.records = l.records[len(l.records)-l.maxSize:]
	}
}

// Recent returns the last n records.
func (l *Logger) Recent(n int) []Record {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if n > len(l.records) {
		n = len(l.records)
	}
	start := len(l.records) - n
	result := make([]Record, n)
	copy(result, l.records[start:])
	return result
}

// All returns all records.
func (l *Logger) All() []Record {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]Record, len(l.records))
	copy(result, l.records)
	return result
}

// ToJSON serializes the log to JSON.
func (l *Logger) ToJSON() ([]byte, error) {
	records := l.All()
	return json.MarshalIndent(records, "", "  ")
}

// Clear empties the log.
func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.records = l.records[:0]
}

// Count returns the number of records.
func (l *Logger) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.records)
}
