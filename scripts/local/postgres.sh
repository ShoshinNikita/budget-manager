#!/bin/bash

docker stop postgres

PWD=$(pwd)
docker run --rm -d \
	--name postgres \
	-p "5432:5432" \
	-v "$PWD/var/pg-data:/var/lib/postgresql/data" \
	-e POSTGRES_USER=postgres \
	-e POSTGRES_DB=postgres \
	postgres -c "log_statement=all"
