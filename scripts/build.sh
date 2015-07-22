#!/bin/bash
cd "${WORKSPACE}/${REPO}"

mkdir ./bin
mkdir ${BUILDDIR}/${REPO}

curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer -o ./gvm-installer.sh
bash ./gvm-installer.sh master gvm
source gvm/gvm/scripts/gvm

gvm install go1.4
gvm use go1.4

export GOPATH=`pwd`

gvm linkthis

go get ./... -v
go build -o ${BUILDDIR}/${REPO}/logdriver -v

gvm implode
