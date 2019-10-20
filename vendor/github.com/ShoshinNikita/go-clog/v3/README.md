# colored-log

Pretty colored logger for Golang

## Preview

[![example](example/example.png)](example/example.go)

## Install

`go get github.com/ShoshinNikita/go-clog/v3`

## Example

[Example program](example/example.go)

```go
package main

import (
    clog "github.com/ShoshinNikita/go-clog/v3"
)

func main() {
    // For prod use log.NewProdConfig() or log.NewProdLogger()
    // For dev use log.NewDevConfig() or log.NewDevLogger()

    c := &clog.Config{}
    l := c.PrintColor(true).PrintErrorLine(true).PrintTime(true).SetLevel(clog.LevelDebug).SetPrefix("prefix: ").Build()
    l.Debug("some debug message")
    l.Info("some info message")
    l.Warn("some warn message")
    l.Error("some error message")

    l.WriteString("\n")

    c = &clog.Config{}
    l = c.PrintColor(true).PrintErrorLine(false).PrintTime(false).SetLevel(clog.LevelDebug).Build()
    l.Debug("some debug message")
    l.Info("some info message")
    l.Warn("some warn message")
    l.Error("some error message")

    l.WriteString("\n")

    c = &clog.Config{}
    l = c.PrintColor(false).PrintErrorLine(false).PrintTime(false).SetLevel(clog.LevelDebug).Build()
    l.Debug("some debug message")
    l.Info("some info message")
    l.Warn("some warn message")
    l.Error("some error message")

    l.WriteString("\n")

    c = &clog.Config{}
    l = c.PrintColor(false).PrintErrorLine(false).PrintTime(false).SetLevel(clog.LevelInfo).Build()
    l.Debug("some debug message")
    l.Info("some info message")
    l.Warn("some warn message")
    l.Error("some error message")
}
```
