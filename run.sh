#!/bin/sh
FILENAME=${1%.*}
BASENAME=${FILENAME##*/}
cat $1
./compile -if=false $1 &&
arm-linux-gnueabi-gcc -o $BASENAME -mcpu=arm1176jzf-s -mtune=arm1176jzf-s $FILENAME.s &&
qemu-arm -L /usr/arm-linux-gnueabi $BASENAME
