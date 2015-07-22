#!/bin/bash
cd "${WORKSPACE}/${REPO}"


GOLANG_TAG="go1.4.1"
INSTALL_DIR=./go-builds/$GOLANG_TAG
mkdir -p $INSTALL_DIR
git clone https://go.googlesource.com/go $INSTALL_DIR
cd $INSTALL_DIR
git checkout $GOLANG_TAG
cd src
bash ./all.bash

mkdir ${BUILDDIR}/${REPO}

export GOPATH=`pwd`

$INSTALL_DIR/bin/go get ./... -v
$INSTALL_DIR/bin/go build -o ${BUILDDIR}/${REPO}/logdriver -v
