function install_golang {
    GOLANG_TAG=${$1:"go1.4.1"}
    INSTALL_DIR=$HOME/go-builds/$GOLANG_TAG
    mkdir -p $INSTALL_DIR
    git clone https://go.googlesource.com/go $INSTALL_DIR
    cd $INSTALL_DIR
    git checkout $GOLANG_TAG
    cd src
    bash ./all.bash
}
