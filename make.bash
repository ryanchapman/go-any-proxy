#!/bin/bash

# Required if you are using a 32-bit vyatta. Otherwise, comment it out
GOOS=linux
#GOARCH=386

function makeVersion()
{
cat <<vEOF >version.go
package main

const BUILDTIMESTAMP = `date +%s`
const BUILDUSER      = "`id -un`"
const BUILDHOST      = "`hostname`"
vEOF
}

function build ()
{
    makeVersion
    go build any_proxy.go stats.go version.go
    return $?
}

case $1 in 
  "clean")
    rm -f version.go
    ;;
  "deploy")
    build && scp any_proxy sapphire:/nfs/local/linux/any_proxy/
    ;;
  *)
    build
    ;;
esac

