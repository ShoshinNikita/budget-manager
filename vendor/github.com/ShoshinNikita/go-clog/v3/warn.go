package clog

import (
	"fmt"
	"time"
)

// Warn prints warning
// Output pattern: (?time) [WRN] (?custom prefix) warning
func (l Logger) Warn(v ...interface{}) {
	print := func() (int, error) {
		return fmt.Fprintln(l.buff, v...)
	}

	l.warn(print)
}

// Warnf prints warning. "\n" is added automatically
// Output pattern: (?time) [WRN] (?custom prefix) warning
func (l Logger) Warnf(format string, v ...interface{}) {
	print := func() (int, error) {
		return fmt.Fprintf(l.buff, format+"\n", v...)
	}

	l.warn(print)
}

// warn is an internal function for printing warning messages
// Output pattern: (?time) [WRN] (?custom prefix) warning
func (l Logger) warn(print messagePrintFunction) {
	if !l.shouldPrint(LevelWarn) {
		return
	}

	now := time.Now()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.buff.Reset()

	l.writeIntoBuffer(l.getTime(now))
	l.writeIntoBuffer(l.getWarnPrefix())
	l.writeIntoBuffer(l.getCustomPrefix())

	print()

	l.output.Write(l.buff.Bytes())
}
