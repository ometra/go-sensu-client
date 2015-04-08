#!/bin/sh
export GOPATH=`pwd`

# set DEBUG to a non empty value to display debug information
export DEBUG=1
go run -a src/sensu-client.go --config-file=src/config/config.json
