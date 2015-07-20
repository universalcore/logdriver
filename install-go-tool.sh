#/bin/bash

set -e

go get -v -u golang.org/x/tools/cmd/goimports
go get -v -u github.com/golang/lint/golint
go get -v -u github.com/nsf/gocode
go get -v -u code.google.com/p/rog-go/exp/cmd/godef
go get -v -u sourcegraph.com/sqs/goreturns
go get -v -u golang.org/x/tools/cmd/vet
go get -v -u golang.org/x/tools/cmd/oracle
go get -v -u github.com/ActiveState/tail
