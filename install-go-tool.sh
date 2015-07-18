#/bin/bash

set -e

GOPATH=`pwd`
go get -u golang.org/x/tools/cmd/goimports
go get -u github.com/golang/lint/golint
go get -u github.com/nsf/gocode
go get -u code.google.com/p/rog-go/exp/cmd/godef
go get -u sourcegraph.com/sqs/goreturns
go get -u golang.org/x/tools/cmd/vet
go get -u golang.org/x/tools/cmd/oracle
