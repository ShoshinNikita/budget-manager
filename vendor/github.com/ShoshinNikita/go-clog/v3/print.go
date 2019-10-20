package clog

import (
	"fmt"
	"time"
)

// Write writes the content of p
func (l *Logger) Write(b []byte) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.output.Write(b)
}

// WriteString writes the content of s
func (l *Logger) WriteString(s string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.output.Write([]byte(s))
}

// Print prints msg
// Output pattern: (?time) (?custom prefix) msg
func (l Logger) Print(v ...interface{}) {
	print := func() (int, error) {
		return fmt.Fprintln(l.buff, v...)
	}

	l.print(print)
}

// Printf prints msg."\n" is added automatically
// Output pattern: (?time) (?custom prefix) msg
func (l Logger) Printf(format string, v ...interface{}) {
	print := func() (int, error) {
		return fmt.Fprintf(l.buff, format+"\n", v...)
	}

	l.print(print)
}

func (l Logger) print(print messagePrintFunction) {
	now := time.Now()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.buff.Reset()

	l.writeIntoBuffer(l.getTime(now))
	l.writeIntoBuffer(l.getCustomPrefix())

	print()

	l.output.Write(l.buff.Bytes())
}
