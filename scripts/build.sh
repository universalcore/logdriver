#!/bin/bash
cd "${WORKSPACE}/${REPO}"

source ./install_golang.sh
install_golang go1.4.1

mkdir ${BUILDDIR}/${REPO}

export GOPATH=`pwd`

$HOME/bin/go get ./... -v
$HOME/bin/go build -o ${BUILDDIR}/${REPO}/logdriver -v
