#!/bin/bash
cd "${WORKSPACE}/${REPO}"
export GOPATH=`pwd`

mkdir ./bin
mkdir ${BUILDDIR}/${REPO}

curl -sL -o ./bin/gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme
chmod +x ./bin/gimme

eval "$(./bin/gimme 1.4)"

go get ./... -v
go build -o ${BUILDDIR}/${REPO}/logdriver -v
