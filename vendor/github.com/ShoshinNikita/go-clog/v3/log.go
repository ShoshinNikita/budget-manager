// Package log provides functions for pretty print
//
// Patterns of functions print:
//
// * (?time) msg - Print(), Printf(), Println():
//
// * (?time) [DBG] msg - Debug(), Debugf(), Debugln():
//
// * (?time) [INF] msg - Info(), Infof(), Infoln():
//
// * (?time) [WRN] warning - Warn(), Warnf(), Warnln():
//
// * (?time) [ERR] (?file:line) error - Error(), Errorf(), Errorln():
//
// * (?time) [FAT] (?file:line) error - Fatal(), Fatalf(), Fatalln():
//
// Time pattern: MM.dd.yyyy hh:mm:ss (01.30.2018 05:5:59)
//
package clog

import (
	"bytes"
	"io"
	"os"
	"sync"

	"github.com/fatih/color"
)

const (
	DefaultTimeLayout = "01.02.2006 15:04:05"
)

type LogLevel int

const (
	LevelDebug = iota - 1
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

type messagePrintFunction func() (int, error)

// -----------------------------------------------------------------------------
// Logger
// -----------------------------------------------------------------------------

type Logger struct {
	output io.Writer
	mutex  *sync.Mutex
	buff   *bytes.Buffer

	level LogLevel

	customPrefix []byte

	printColor     bool
	printErrorLine bool

	printTime  bool
	timeLayout string
}

func NewDevLogger() *Logger {
	return NewDevConfig().Build()
}

func NewProdLogger() *Logger {
	return NewProdConfig().Build()
}

// WithPrefix returns cloned Logger with added prefix (prefix + ": ").
//
// Example:
//   - l.prefix == "old prefix"
//   - l.WithPrefix("new prefix")
//   - l.prefix == "new prefix: old prefix"
//
func (l Logger) WithPrefix(prefix string) *Logger {
	if prefix == "" {
		return l.clone()
	}

	log := l.clone()

	prefix = prefix + ": "
	log.customPrefix = append([]byte(prefix), log.customPrefix...)

	return log
}

// -----------------------------------------------------------------------------
// Config
// -----------------------------------------------------------------------------

type Config struct {
	output io.Writer

	level LogLevel

	prefix []byte

	printColor     bool
	printErrorLine bool

	printTime  bool
	timeLayout string
}

func NewDevConfig() *Config {
	return &Config{
		output:         color.Output,
		level:          LevelDebug,
		prefix:         nil,
		printTime:      true,
		printColor:     true,
		printErrorLine: true,
		timeLayout:     DefaultTimeLayout,
	}
}

func NewProdConfig() *Config {
	return &Config{
		output:         os.Stdout,
		level:          LevelInfo,
		prefix:         nil,
		printTime:      true,
		printColor:     false,
		printErrorLine: true,
		timeLayout:     DefaultTimeLayout,
	}
}

// Build create a new Logger according to Config
func (c *Config) Build() *Logger {
	l := new(Logger)
	l.mutex = new(sync.Mutex)
	l.buff = new(bytes.Buffer)

	switch {
	case c.printColor && c.output == nil:
		l.output = color.Output
	case c.output != nil:
		l.output = c.output
	default:
		l.output = os.Stdout
	}

	l.level = c.level
	l.customPrefix = c.prefix
	l.printTime = c.printTime
	l.printColor = c.printColor
	l.printErrorLine = c.printErrorLine

	l.timeLayout = DefaultTimeLayout
	if c.timeLayout != "" {
		l.timeLayout = c.timeLayout
	}

	return l
}

// Debug sets Config.debug to b
func (c *Config) SetLevel(lvl LogLevel) *Config {
	c.level = lvl
	return c
}

// PrintTime sets Config.printTime to b
func (c *Config) PrintTime(b bool) *Config {
	c.printTime = b
	return c
}

// PrintColor sets Config.printColor to b
func (c *Config) PrintColor(b bool) *Config {
	c.printColor = b
	return c
}

// PrintErrorLine sets Config.printErrorLine to b
func (c *Config) PrintErrorLine(b bool) *Config {
	c.printErrorLine = b
	return c
}

// SetOutput changes Config.output writer.
func (c *Config) SetOutput(w io.Writer) *Config {
	c.output = w
	return c
}

// SetTimeLayout changes Config.timeLayout
// Default Config.timeLayout is DefaultTimeLayout
func (c *Config) SetTimeLayout(layout string) *Config {
	c.timeLayout = layout
	return c
}

// SetPrefix overwrites Config.prefix. It doesn't add ": "!
func (c *Config) SetPrefix(prefix string) *Config {
	c.prefix = []byte(prefix)
	return c
}
