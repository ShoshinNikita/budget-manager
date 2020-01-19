module github.com/ShoshinNikita/budget-manager

go 1.13

require (
	github.com/abbot/go-http-auth v0.4.0
	github.com/caarlos0/env/v6 v6.1.0
	github.com/fatih/color v1.9.0
	github.com/go-pg/pg/v9 v9.1.1
	github.com/go-pg/urlstruct v0.2.11 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/segmentio/encoding v0.1.9 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/vmihailenco/msgpack/v4 v4.3.5 // indirect
	golang.org/x/crypto v0.0.0-20200117160349-530e935923ad // indirect
	golang.org/x/net v0.0.0-20200114155413-6afb5195e5aa // indirect
	golang.org/x/sys v0.0.0-20200117145432-59e60aa80a0c // indirect
)

// Commit with the license
replace github.com/go-pg/zerochecker => github.com/go-pg/zerochecker v0.1.2-0.20190924102134-549b3bbf317c
