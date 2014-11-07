#!/bin/bash
export DEBIAN_FRONTEND=noninteractive
if [ -n "`which apt-get`" ];
then
    mkdir $HOME/go
    echo 'export GOPATH=$HOME/go' >> ~/.bash_profile
    echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bash_profile
    echo 'export GOROOT=/usr/lib/go' >> ~/.bash_profile
    apt-get install -y golang-go
    source ~/.bash_profile
    go get github.com/blynn/nex
fi
