#!/bin/sh
TMP=/tmp
FULLFN=$(readlink -f $1)
FILENAME=${FULLFN%.*}
COMPILE=$(readlink -f ./compile)
BASENAME=${FILENAME##*/}

cat $FULLFN
$COMPILE -o $TMP/$BASENAME.s $FULLFN &&
arm-linux-gnueabi-gcc -o $TMP/$BASENAME -mcpu=arm1176jzf-s -mtune=arm1176jzf-s $TMP/$BASENAME.s &&
qemu-arm -L /usr/arm-linux-gnueabi $TMP/$BASENAME
