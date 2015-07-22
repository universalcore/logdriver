#!/bin/bash
REPO_DIR="$WORKSPACE/$REPO"
GOLANG_TAG="go1.4.1"
INSTALL_DIR=$REPO_DIR/go-builds/$GOLANG_TAG
if [ ! -d $INSTALL_DIR ]; then
    mkdir -p $INSTALL_DIR
    git clone https://go.googlesource.com/go $INSTALL_DIR
    cd $INSTALL_DIR
    git checkout $GOLANG_TAG
    cd src
    bash ./make.bash
fi

mkdir ${BUILDDIR}/${REPO}

export GOPATH=`pwd`

cd $REPO_DIR
$INSTALL_DIR/bin/go get github.com/ActiveState/tail
$INSTALL_DIR/bin/go get github.com/gorilla/mux
$INSTALL_DIR/bin/go build -o ${BUILDDIR}/${REPO}/logdriver -v
