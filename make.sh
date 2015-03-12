#!/bin/bash
set -e
cd `dirname $0`
tar czf ui/embed/module.tar.gz *.c *.h Makefile Kbuild
cd ui
go get github.com/GeertJohan/go.rice/...
rice clean
rice embed-go
go build
mv ui ../gueststack
