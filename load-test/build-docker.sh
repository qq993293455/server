#!/usr/bin/env bash

workdir=$(cd ../ && pwd)
cd $workdir
echo "go build ..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o ./load-test/load-test ./load-test/main.go
cd $workdir/load-test
docker build -t load-test .