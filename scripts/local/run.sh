#!/bin/bash

export DEBUG="true"
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="postgres"
export DB_DATABASE="postgres"
export SERVER_PORT="8080"
export SERVER_SKIP_AUTH="true"
# user:qwerty,admin:admin
# export SERVER_CREDENTIALS="user:\$apr1\$cpHMFyv.\$BSB0aaF3bOrTC2f3V2VYG/,admin:\$apr1\$t6YLYGF6\$M05uLevUvoHOopO6AUOEj/"

go run -mod vendor cmd/budget-manager/main.go
