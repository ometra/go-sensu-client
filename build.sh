#!/usr/bin/env bash

export NDK_TOOLCHAIN=~/dev/ndk-toolchain
export CC=$NDK_TOOLCHAIN/bin/arm-linux-androideabi-gcc
export GOROOT=~/dev/go
export GOPATH=`pwd`
export GOOS=android
export GOARCH=arm
export GOARM=7
export CGO_ENABLED=1

GO="$GOROOT/bin/go"
LD_FLAGS=''
#LD_FLAGS='-ldflags="-shared"' # for building a shared library for java

echo "Bulding with..."
$GO version

#echo "Getting Dependencies..."
#$GO get golang.org/x/mobile/app

echo "Building android/arm7 Client..."
$GO build $LD_FLAGS -o bin/sensu-client-arm src/sensu-client.go src/sensu-client_android.go

echo "Building Linux/x64 Client..."
export GOOS=linux
export GOARCH=amd64
$GO build -o bin/sensu-client-x64 src/sensu-client.go
$GO build -o bin/check-procs src/check-procs.go

echo "Done. The binaries are in ./bin/"
