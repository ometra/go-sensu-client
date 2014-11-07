#!/bin/sh
export GOPATH=`pwd`

# set DEBUG to a non empty value to display debug information
export DEBUG=
go run src/sensu-client.go --config-file=src/config/config.json
