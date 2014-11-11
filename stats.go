//
// any_proxy.go - Transparently proxy a connection using Linux iptables REDIRECT
//
// Copyright (C) 2013 Ryan A. Chapman. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
//   1. Redistributions of source code must retain the above copyright notice,
//      this list of conditions and the following disclaimer.
//
//   2. Redistributions in binary form must reproduce the above copyright notice,
//      this list of conditions and the following disclaimer in the documentation
//      and/or other materials provided with the distribution.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
// INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND
// FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE AUTHORS
// OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
// EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS;
// OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY,
// WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR
// OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF
// ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// Ryan A. Chapman, ryan@rchapman.org
// Sun Apr  7 21:04:34 MDT 2013
//


package main

import (
    "fmt"
    log "github.com/zdannar/flogger"
    "os"
    "os/signal"
    "runtime"
    "sync"
    "syscall"
    "time"
)

var acceptErrors struct {
    sync.Mutex
    n uint64
}

var acceptSuccesses struct {
    sync.Mutex
    n uint64
}

var getOriginalDstErrors struct {
    sync.Mutex
    n uint64
}

var directConnections struct {
    sync.Mutex
    n uint64
}

var proxiedConnections struct {
    sync.Mutex
    n uint64
}

var proxy200Responses struct {
    sync.Mutex
    n uint64
}

var proxy300Responses struct {
    sync.Mutex
    n uint64
}

var proxy400Responses struct {
    sync.Mutex
    n uint64
}

var proxyNon200Responses struct {
    sync.Mutex
    n uint64
}

var proxyNoConnectResponses struct {
    sync.Mutex
    n uint64
}

var proxyServerReadErr struct {
    sync.Mutex
    n uint64
}

var proxyServerWriteErr struct {
    sync.Mutex
    n uint64
}

var directServerReadErr struct {
    sync.Mutex
    n uint64
}

var directServerWriteErr struct {
    sync.Mutex
    n uint64
}

func incrAcceptErrors() {
    acceptErrors.Lock()
    acceptErrors.n++
    acceptErrors.Unlock()
}

func numAcceptErrors() (uint64) {
    return acceptErrors.n
}

func incrAcceptSuccesses() {
    acceptSuccesses.Lock()
    acceptSuccesses.n++
    acceptSuccesses.Unlock()
}

func numAcceptSuccesses() (uint64) {
    return acceptSuccesses.n
}

func incrGetOriginalDstErrors() {
    getOriginalDstErrors.Lock()
    getOriginalDstErrors.n++
    getOriginalDstErrors.Unlock()
}

func numGetOriginalDstErrors() (uint64) {
    return getOriginalDstErrors.n
}

func incrDirectConnections() {
    directConnections.Lock()
    directConnections.n++
    directConnections.Unlock()
}

func numDirectConnections() (uint64) {
    return directConnections.n
}

func incrProxiedConnections() {
    proxiedConnections.Lock()
    proxiedConnections.n++
    proxiedConnections.Unlock()
}

func numProxiedConnections() (uint64) {
    return proxiedConnections.n
}

func incrProxy200Responses() {
    proxy200Responses.Lock()
    proxy200Responses.n++
    proxy200Responses.Unlock()
}

func numProxy200Responses() (uint64) {
    return proxy200Responses.n
}

func incrProxy300Responses() {
    proxy300Responses.Lock()
    proxy300Responses.n++
    proxy300Responses.Unlock()
}

func numProxy300Responses() (uint64) {
    return proxy300Responses.n
}

func incrProxy400Responses() {
    proxy400Responses.Lock()
    proxy400Responses.n++
    proxy400Responses.Unlock()
}

func numProxy400Responses() (uint64) {
    return proxy400Responses.n
}

func incrProxyNon200Responses() {
    proxyNon200Responses.Lock()
    proxyNon200Responses.n++
    proxyNon200Responses.Unlock()
}

func numProxyNon200Responses() (uint64) {
    return proxyNon200Responses.n
}

func incrProxyNoConnectResponses() {
    proxyNoConnectResponses.Lock()
    proxyNoConnectResponses.n++
    proxyNoConnectResponses.Unlock()
}

func numProxyNoConnectResponses() (uint64) {
    return proxyNoConnectResponses.n
}

func incrProxyServerReadErr() {
    proxyServerReadErr.Lock()
    proxyServerReadErr.n++
    proxyServerReadErr.Unlock()
}

func numProxyServerReadErr() (uint64) {
    return proxyServerReadErr.n
}

func incrProxyServerWriteErr() {
    proxyServerWriteErr.Lock()
    proxyServerWriteErr.n++
    proxyServerWriteErr.Unlock()
}

func numProxyServerWriteErr() (uint64) {
    return proxyServerWriteErr.n
}

func incrDirectServerReadErr() {
    directServerReadErr.Lock()
    directServerReadErr.n++
    directServerReadErr.Unlock()
}

func numDirectServerReadErr() (uint64) {
    return directServerReadErr.n
}

func incrDirectServerWriteErr() {
    directServerWriteErr.Lock()
    directServerWriteErr.n++
    directServerWriteErr.Unlock()
}

func numDirectServerWriteErr()(uint64) {
    return directServerWriteErr.n
}

func setupStats() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGUSR1)
    go func() {
        for _ = range c {
            f, err := os.Create(STATSFILE)
            if err != nil {
                log.Infof("ERR: Could not open stats file \"%s\": %v", STATSFILE, err)
                continue
            }
            fmt.Fprintf(f, "%s\n\n", versionString())
            fmt.Fprintf(f, "STATISTICS as of %v:\n", time.Now().Format(time.UnixDate))
            fmt.Fprintf(f, "                                Go version: %v\n", runtime.Version())
            fmt.Fprintf(f, "          Number of logical CPUs on system: %v\n", runtime.NumCPU())
            fmt.Fprintf(f, "                                GOMAXPROCS: %v\n", runtime.GOMAXPROCS(-1))
            fmt.Fprintf(f, "              Goroutines currently running: %v\n", runtime.NumGoroutine())
            fmt.Fprintf(f, "     Number of cgo calls made by any_proxy: %v\n", runtime.NumCgoCall())
            fmt.Fprintf(f, "\n")
            fmt.Fprintf(f, "                          accept successes: %v\n", numAcceptSuccesses())
            fmt.Fprintf(f, "                             accept errors: %v\n", numAcceptErrors())
            fmt.Fprintf(f, "        getsockopt(SO_ORIGINAL_DST) errors: %v\n", numGetOriginalDstErrors())
            fmt.Fprintf(f, "\n")
            fmt.Fprintf(f, "                 connections sent directly: %v\n", numDirectConnections())
            fmt.Fprintf(f, "             direct connection read errors: %v\n", numDirectServerReadErr())
            fmt.Fprintf(f, "            direct connection write errors: %v\n", numDirectServerWriteErr())
            fmt.Fprintf(f, "\n")
            fmt.Fprintf(f, "        connections sent to upstream proxy: %v\n", numProxiedConnections())
            fmt.Fprintf(f, "              proxy connection read errors: %v\n", numProxyServerReadErr())
            fmt.Fprintf(f, "             proxy connection write errors: %v\n", numProxyServerWriteErr())
            fmt.Fprintf(f, "           code 200 response from upstream: %v\n", numProxy200Responses())
            fmt.Fprintf(f, "           code 400 response from upstream: %v\n", numProxy400Responses())
            fmt.Fprintf(f, "other (non 200/400) response from upstream: %v\n", numProxyNon200Responses())
            fmt.Fprintf(f, "      no response to CONNECT from upstream: %v\n", numProxyNoConnectResponses())
            f.Close()
        }
    }()
}


