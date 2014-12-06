#!/bin/sh
TMPO=/tmp
FULLFN=$(readlink -f $1)
FILENAME=${FULLFN%.*}
COMPILE=$(readlink -f ./compile)
BASENAME=${FILENAME##*/}
pushd /tmp >/dev/null 2>&1
cat $FULLFN
$COMPILE -if=false $FULLFN &&
arm-linux-gnueabi-gcc -o $BASENAME -mcpu=arm1176jzf-s -mtune=arm1176jzf-s $BASENAME.s &&
qemu-arm -L /usr/arm-linux-gnueabi $BASENAME
EXIT=$?
popd >/dev/null 2>&1
exit $EXIT
