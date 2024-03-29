#!/bin/bash

if [[ ! $(go env GOOS) == "linux" ]] ; then
    echo "$0: wrong os: Building a Kerlink IoT Station requires a Linux environment. Please retry on another environment."
    exit 1
fi

TOOLCHAIN_PATH="$1"

usage_str="usage: build-kerlink.sh [path to the Kerlink toolchain]
The Kerlink toolchain for the appropriate firmware can be downloaded on the Kerlink Wiki, at
http://wikikerlink.fr/lora-station/doku.php?id=wiki:ressources
at the \"Tools\" section.

example: ./build-kerlink.sh /opt/arm-2011.03-wirgrid"

if [[ -z "$TOOLCHAIN_PATH" ]] ; then
    echo "$0: $usage_str"
    exit 1
fi

pushd "$GOPATH/src/github.com/TheThingsNetwork/packet_forwarder"

export CROSS_COMPILE=arm-none-linux-gnueabi-
export CC=arm-none-linux-gnueabi-gcc
export GOARM=5
export GOOS=linux
export GOARCH=arm
export PLATFORM=kerlink
export GPS_PATH="/dev/nmea"

export PATH="$PATH:$TOOLCHAIN_PATH/bin"

make dev-deps
make deps
make build
if ! make build ; then
    echo "$0: Build of the packet forwarder has failed. Make sure the toolchain is available, and that you have installed the appropriate dependencies (with \`make dev-deps\` and \`make deps\`)."
    exit 1
fi

echo "$0: Build complete and available in $PWD/release."

popd
