module github.com/ShoshinNikita/budget-manager

go 1.13

require (
	github.com/abbot/go-http-auth v0.4.0
	github.com/caarlos0/env/v6 v6.2.2
	github.com/fatih/color v1.9.0
	github.com/go-pg/migrations/v7 v7.1.10
	github.com/go-pg/pg/v9 v9.1.6
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/gorilla/mux v1.7.4
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/segmentio/encoding v0.1.12 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/objx v0.1.1 // indirect
	github.com/stretchr/testify v1.5.1
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37 // indirect
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2 // indirect
	golang.org/x/sys v0.0.0-20200519105757-fe76b779f299 // indirect
)

// Commit with the license
replace github.com/go-pg/zerochecker => github.com/go-pg/zerochecker v0.1.2-0.20190924102134-549b3bbf317c
