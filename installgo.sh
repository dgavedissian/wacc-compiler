export DEBIAN_FRONTEND=noninteractive
if [ -n "`which apt-get`" ];
then
    apt-get install -y golang-go
fi
