#!/bin/sh

GOOS=linux GOARCH=mips CGO_ENABLED=0 GOMIPS=softfloat go build -a -trimpath -ldflags="-s -w"
