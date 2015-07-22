#!/bin/bash
cd "${WORKSPACE}/${REPO}"
export GOPATH=`pwd`

mkdir ./bin
mkdir ${BUILDDIR}/${REPO}

bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
gvm install go1.4
gvm use go1.4

go get ./... -v
go build -o ${BUILDDIR}/${REPO}/logdriver -v

gvm implode
