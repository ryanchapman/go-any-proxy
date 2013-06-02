#!/bin/bash

# Required if you are using a 32-bit vyatta. Otherwise, comment it out
GOOS=linux
#GOARCH=386

function make_version ()
{
    local timestamp=`date +%s`
    local builduser=`id -un`
    local buildhost=`hostname`
cat <<vEOF >$BUILD_DIR/version.go
package main

const BUILDTIMESTAMP = $timestamp
const BUILDUSER      = "$builduser"
const BUILDHOST      = "$buildhost"
vEOF
    echo "Wrote $BUILD_DIR/version.go: timestamp=$timestamp; builduser=$builduser; buildhost=$buildhost"
}

function build ()
{
    make_version
    go build any_proxy.go stats.go version.go
    return $?
}

function build_failed ()
{
    echo "TRAVIS_TEST_RESULT=$TRAVIS_TEST_RESULT"
    echo "Build failed."
    echo "CWD:"
    pwd | sed 's/^/  /g'
    echo "Environment:"
    set | sed 's/^/  /g'
    echo "$BUILD_DIR/version.go:"
    cat $BUILD_DIR/version.go | sed 's/^/  /g'
}


export BUILD_DIR=$2
if [ "$BUILD_DIR" = "" ]; then export BUILD_DIR=.; fi
case $1 in 
  "clean")
    rm -f version.go
    ;;
  "deploy")
    build && scp any_proxy sapphire:/nfs/local/linux/any_proxy/
    ;;
  "version")
    make_version
    ;;
  "build_failed")
    build_failed
    ;;
  *)
    build
    ;;
esac

