module github.com/ShoshinNikita/budget-manager

go 1.13

require (
	github.com/ShoshinNikita/go-clog/v3 v3.2.0
	github.com/abbot/go-http-auth v0.4.0
	github.com/caarlos0/env/v6 v6.1.0
	github.com/fatih/color v1.8.0 // indirect
	github.com/go-pg/pg/v9 v9.1.1
	github.com/go-pg/urlstruct v0.2.10 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.11 // indirect
	github.com/pkg/errors v0.8.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/stretchr/testify v1.4.0
	golang.org/x/crypto v0.0.0-20191227163750-53104e6ec876 // indirect
	golang.org/x/sys v0.0.0-20200106162015-b016eb3dc98e // indirect
)

replace (
	// Commit with the license
	github.com/go-pg/zerochecker => github.com/go-pg/zerochecker v0.1.2-0.20190924102134-549b3bbf317c
	github.com/kr/pretty => github.com/kr/pretty v0.2.0
	github.com/kr/pty => github.com/creack/pty v1.1.9
)
