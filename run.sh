#!/bin/sh
export GOPATH=`pwd`
go run src/sensu-client.go --config-file=src/config/config.json
