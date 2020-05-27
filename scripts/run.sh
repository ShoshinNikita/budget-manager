#!/bin/bash

export DEBUG="true"
export LOGGER_MODE="develop"
export DB_TYPE="postgres"
export DB_PG_HOST="localhost"
export DB_PG_PORT="5432"
export DB_PG_USER="postgres"
export DB_PG_DATABASE="postgres"
export SERVER_PORT="8080"
export SERVER_SKIP_AUTH="true"
export SERVER_CACHE_TEMPLATES="false"
# user:qwerty,admin:admin
# export SERVER_CREDENTIALS="user:\$apr1\$cpHMFyv.\$BSB0aaF3bOrTC2f3V2VYG/,admin:\$apr1\$t6YLYGF6\$M05uLevUvoHOopO6AUOEj/"
export SERVER_ENABLE_PROFILING="true"

go run -ldflags "${LDFLAGS}" -mod vendor cmd/budget-manager/main.go
