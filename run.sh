#!/bin/sh
export GOPATH=/home/jason/swift/go-sensu-client
go run src/sensu-client.go --config-file=src/config/config.json