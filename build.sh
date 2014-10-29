GOPATH=`pwd` GOARCH=arm go build -o bin/sensu-client-arm src/sensu-client.go
GOPATH=`pwd` go build -o bin/sensu-client-x64 src/sensu-client.go
GOPATH=`pwd` go build -o bin/check-procs src/check-procs.go
