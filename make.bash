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

function pull_deps()
{
    go get -u github.com/zdannar/flogger
    go get -u github.com/namsral/flag
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

function make_debian_package ()
{
    local VERSION=`./any_proxy -h | head -1 | awk '{print $2 "-" $4}' | tr -d ','`
    echo "any-proxy ($VERSION) UNRELEASED; urgency=low" > debian/changelog
    git log | sed 's/^/  /g' >> debian/changelog
    echo " -- Ryan A. Chapman <ryan@rchapman.org>  `date '+%a, %d %b %Y %H:%M:%S %z'`" >> debian/changelog
    dpkg-buildpackage -d
}

function reindex_debian_packages ()
{
    echo "Rebuilding debian package repos"
    (cd debian/apt && ./reindex_stable.sh)
}

export BUILD_DIR=$2
if [ "$BUILD_DIR" = "" ]; then export BUILD_DIR=.; fi
case $1 in 
  "clean")
    rm -f version.go any_proxy
    ;;
  "version")
    make_version
    ;;
  "build_failed")
    build_failed
    ;;
  "package_write_pubkey")
    TMPF=`mktemp /tmp/anyproxy_pub.XXX`.key
    gpg --export -a "ryan@rchapman.org" > $TMPF
    echo "Wrote pubkey to $TMPF"
    echo "Importing pubkey $TMPF into debian/apt/repo.gpg"
    gpg --no-default-keyring --keyring debian/apt/repo.gpg --import $TMPF
    if [[ $? == 0 ]]; then
        echo "Wrote pubkey successfully to debian/apt/repo.gpg"
    else
        echo "ERROR: could not write pubkey to debian/apt/repo.gpg"
    fi
    rm -f $TMPF
    ;;
  "package")
    if [[ -f any_proxy ]]; then
        echo -n "A built product already exist. Re-packaging existing build (y/n) [y]: "
        read ANS
        if [[ $ANS != n && $ANS != N ]]; then
            make_debian_package
            reindex_debian_packages
        fi
    fi
    ;;
  *)
    pull_deps
    build
    ;;
esac

