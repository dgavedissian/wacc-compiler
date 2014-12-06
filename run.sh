#!/bin/sh
TMPO=/tmp
FULLFN=$(readlink -f $1)
FILENAME=${FULLFN%.*}
COMPILE=$(readlink -f ./compile)
BASENAME=${FILENAME##*/}
pushd /tmp
cat $FULLFN
$COMPILE -if=false $FULLFN &&
arm-linux-gnueabi-gcc -o $BASENAME -mcpu=arm1176jzf-s -mtune=arm1176jzf-s $BASENAME.s &&
qemu-arm -L /usr/arm-linux-gnueabi $BASENAME
popd
