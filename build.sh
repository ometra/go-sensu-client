#!/usr/bin/env bash

echo "Building android/arm7 Client"
GOROOT=~/dev/go
GOPATH=`pwd`
ARM_GOARCH=arm
ARM_GOOS=android
ARM_GOARM=7
GO="$GOROOT/bin/go"
LD_FLAGS=''
#LD_FLAGS='-ldflags="-shared"'

GOROOT=$GOROOT GOPATH=$GOPATH GOOS=$ARM_GOOS GOARCH=$ARM_GOARCH GOARM=$ARM_GOARM CGO_ENABLED=1 $GO build $LD_FLAGS -o bin/sensu-client-arm src/sensu-client.go

echo "Building Linux/x64 Client..."
GOROOT=$GOROOT GOPATH=$GOPATH $GO build -o bin/sensu-client-x64 src/sensu-client.go
#GOPATH=`pwd` go build -o bin/check-procs src/check-procs.go

echo "Done. The binaries are in ./bin/"
