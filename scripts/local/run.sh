#!/bin/bash

export DEBUG="true"
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="postgres"
export DB_DATABASE="postgres"
export SERVER_PORT="8080"

go run -mod vendor cmd/budget_manager/main.go
