#!/bin/bash
export DEBIAN_FRONTEND=noninteractive
if [ -n "`which apt-get`" ] && [ -z "`which go`" ];
then
    mkdir $HOME/go
    apt-get install -y golang-go
fi
