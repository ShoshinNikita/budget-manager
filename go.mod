module github.com/ShoshinNikita/budget-manager

go 1.13

require (
	github.com/abbot/go-http-auth v0.4.0
	github.com/caarlos0/env/v6 v6.2.1
	github.com/fatih/color v1.9.0
	github.com/go-pg/migrations/v7 v7.1.9
	github.com/go-pg/pg/v9 v9.1.3
	github.com/golang/protobuf v1.3.4 // indirect
	github.com/gorilla/mux v1.7.4
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect
)

// Commit with the license
replace github.com/go-pg/zerochecker => github.com/go-pg/zerochecker v0.1.2-0.20190924102134-549b3bbf317c
