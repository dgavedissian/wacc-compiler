export DEBIAN_FRONTEND=noninteractive
if [ -n "`which apt-get`" ];
then
    mkdir $HOME/go
    export GOPATH=$HOME/go
    export PATH=$PATH:$GOPATH/bin
    export GOROOT=/usr/lib/go
    apt-get install -y golang-go
fi
