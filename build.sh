#!/bin/bash

EXTENSION=""
ARCH="$(go env GOHOSTARCH)"
DISTPATH="bin/linux-${ARCH}"
if [ "$(go env GOOS)" = "windows" ]; then
    EXTENSION=".exe"
    DISTPATH="bin\\windows-${ARCH}"
fi

go build -o $DISTPATH/smtpsend$EXTENSION github.com/zerklabs/smtpsend
