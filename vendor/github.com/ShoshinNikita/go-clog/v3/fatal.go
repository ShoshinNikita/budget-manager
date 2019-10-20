package clog

import (
	"fmt"
	"os"
	"time"
)

// Fatal prints error and call os.Exit(1)
// Output pattern: (?time) [FAT] (?file:line) (?custom prefix) error
func (l Logger) Fatal(v ...interface{}) {
	print := func() (int, error) {
		return fmt.Fprintln(l.buff, v...)
	}

	l.fatal(print)
}

// Fatalf prints error and call os.Exit(1). "\n" is added automatically
// Output pattern: (?time) [FAT] (?file:line) (?custom prefix) error
func (l Logger) Fatalf(format string, v ...interface{}) {
	print := func() (int, error) {
		return fmt.Fprintf(l.buff, format+"\n", v...)
	}

	l.fatal(print)
}

// fatal is an internal function for printing fatal messages. Is also calls os.Exit(1)
// Output pattern: (?time) [FAT] (?file:line) (?custom prefix) error
func (l Logger) fatal(print messagePrintFunction) {
	if !l.shouldPrint(LevelFatal) {
		os.Exit(1)
		return
	}

	now := time.Now()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.buff.Reset()

	l.writeIntoBuffer(l.getTime(now))
	l.writeIntoBuffer(l.getFatalPrefix())
	l.writeIntoBuffer(l.getCaller())
	l.writeIntoBuffer(l.getCustomPrefix())

	print()

	l.output.Write(l.buff.Bytes())

	os.Exit(1)
}
