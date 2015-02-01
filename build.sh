#!/usr/bin/env bash

export CC=${NDK_TOOLCHAIN:="~/dev/ndk-toolchain"}/bin/arm-linux-androideabi-gcc
export GOROOT=${GOROOT:="/usr/local/go"}
export GOPATH=`pwd`

GO="$GOROOT/bin/go"
BUILD_FLAGS=''

echo "Bulding with..."
${GO} version

if [ -f "$CC" ]; then
    echo "Building android/arm7 Client..."
    export GOOS=android
    export GOARCH=arm
    export GOARM=7
    export CGO_ENABLED=1
    ${GO} build ${BUILD_FLAGS} -o bin/sensu-client-armv7-android src/sensu-client.go src/sensu-client_android.go
else
    echo "No android build environment. skipping android builds..."
fi

export GOOS=linux
export GOARCH=arm
export CGO_ENABLED=0
export CC=""
BUILD_FLAGS="-compiler=gc"

echo "Building Linux/armv7 and Linux/armv6 Clients..."
${GO} build ${BUILD_FLAGS} -o bin/sensu-client-armv7-linux src/sensu-client.go
export GOARM=6
${GO} build ${BUILD_FLAGS} -o bin/sensu-client-armv6-linux src/sensu-client.go


echo "Building Linux/x64 Client..."
export GOOS=linux
export GOARM=""
export GOARCH=amd64
export CGO_ENABLED=1
${GO} build -o bin/sensu-client-x64-linux src/sensu-client.go
${GO} build -o bin/check-procs-x64-linux src/check-procs.go
${GO} build -o bin/metric-tcp-x64-linux src/metric-tcp.go

echo "Done. The binaries are in ./bin/"
