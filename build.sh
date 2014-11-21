#!/usr/bin/env bash


GOROOT=~/dev/go
GOPATH=`pwd`
GOARCH=arm
GOOS=android
GOARM=7
GO="$GOROOT/bin/go"
LD_FLAGS=''
#LD_FLAGS='-ldflags="-shared"'

GOROOT=$GOROOT GOPATH=$GOPATH GOARCH=$GOARCH GOOS=$GOOS GOARM=$GOARM CGO_ENABLED=1 $GO build $LD_FLAGS -o bin/sensu-client-arm src/sensu-client.go
#GOPATH=`pwd` go build -o bin/sensu-client-x64 src/sensu-client.go
#GOPATH=`pwd` go build -o bin/check-procs src/check-procs.go

echo "Done. The binaries are in ./bin/"
