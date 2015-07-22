#!/bin/bash
cd "${WORKSPACE}/${REPO}"
export GOPATH=`pwd`
mkdir ${BUILDDIR}/${REPO}
go build -o ${BUILDDIR}/${REPO}/logdriver -v
