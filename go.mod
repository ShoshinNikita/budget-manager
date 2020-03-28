module github.com/ShoshinNikita/budget-manager

go 1.13

require (
	github.com/abbot/go-http-auth v0.4.0
	github.com/caarlos0/env/v6 v6.2.1
	github.com/fatih/color v1.9.0
	github.com/go-pg/migrations/v7 v7.1.9
	github.com/go-pg/pg/v9 v9.1.5
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/gorilla/mux v1.7.4
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.5.0
	github.com/stretchr/objx v0.1.1 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/vmihailenco/msgpack/v4 v4.3.11 // indirect
	golang.org/x/crypto v0.0.0-20200323165209-0ec3e9974c59 // indirect
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/sys v0.0.0-20200327173247-9dae0f8f5775 // indirect
)

// Commit with the license
replace github.com/go-pg/zerochecker => github.com/go-pg/zerochecker v0.1.2-0.20190924102134-549b3bbf317c
