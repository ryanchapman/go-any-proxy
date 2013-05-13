package main

import (
    "fmt"
    log "./flogger"
    "os"
    "os/signal"
    "runtime"
    "sync"
    "syscall"
    "time"
)

/* STATS */
var acceptErrorsMu             sync.Mutex
var acceptErrors               uint64 = 0

var acceptSuccessesMu          sync.Mutex
var acceptSuccesses            uint64 = 0

var getOriginalDstErrorsMu     sync.Mutex
var getOriginalDstErrors       uint64 = 0

var directConnectionsMu        sync.Mutex
var directConnections          uint64 = 0

var proxiedConnectionsMu       sync.Mutex
var proxiedConnections         uint64 = 0

var proxy200ResponsesMu        sync.Mutex
var proxy200Responses          uint64 = 0

var proxy400ResponsesMu        sync.Mutex
var proxy400Responses          uint64 = 0

var proxyNon200ResponsesMu     sync.Mutex
var proxyNon200Responses       uint64 = 0

var proxyNoConnectResponsesMu  sync.Mutex
var proxyNoConnectResponses    uint64 = 0

var proxyServerReadErrMu       sync.Mutex
var proxyServerReadErr         uint64 = 0

var proxyServerWriteErrMu      sync.Mutex
var proxyServerWriteErr        uint64 = 0

var directServerReadErrMu      sync.Mutex
var directServerReadErr        uint64 = 0

var directServerWriteErrMu     sync.Mutex
var directServerWriteErr       uint64 = 0

func incrAcceptErrors() {
    acceptErrorsMu.Lock()
    acceptErrors += 1
    acceptErrorsMu.Unlock()
}

func numAcceptErrors() (uint64) {
    return acceptErrors
}

func incrAcceptSuccesses() {
    acceptSuccessesMu.Lock()
    acceptSuccesses += 1
    acceptSuccessesMu.Unlock()
}

func numAcceptSuccesses() (uint64) {
    return acceptSuccesses
}

func incrGetOriginalDstErrors() {
    getOriginalDstErrorsMu.Lock()
    getOriginalDstErrors += 1
    getOriginalDstErrorsMu.Unlock()
}

func numGetOriginalDstErrors() (uint64) {
    return getOriginalDstErrors
}

func incrDirectConnections() {
    directConnectionsMu.Lock()
    directConnections += 1
    directConnectionsMu.Unlock()
}

func numDirectConnections() (uint64) {
    return directConnections
}

func incrProxiedConnections() {
    proxiedConnectionsMu.Lock()
    proxiedConnections += 1
    proxiedConnectionsMu.Unlock()
}

func numProxiedConnections() (uint64) {
    return proxiedConnections
}

func incrProxy200Responses() {
    proxy200ResponsesMu.Lock()
    proxy200Responses += 1
    proxy200ResponsesMu.Unlock()
}

func numProxy200Responses() (uint64) {
    return proxy200Responses
}

func incrProxy400Responses() {
    proxy400ResponsesMu.Lock()
    proxy400Responses += 1
    proxy400ResponsesMu.Unlock()
}

func numProxy400Responses() (uint64) {
    return proxy400Responses
}

func incrProxyNon200Responses() {
    proxyNon200ResponsesMu.Lock()
    proxyNon200Responses += 1
    proxyNon200ResponsesMu.Unlock()
}

func numProxyNon200Responses() (uint64) {
    return proxyNon200Responses
}

func incrProxyNoConnectResponses() {
    proxyNoConnectResponsesMu.Lock()
    proxyNoConnectResponses += 1
    proxyNoConnectResponsesMu.Unlock()
}

func numProxyNoConnectResponses() (uint64) {
    return proxyNoConnectResponses
}

func incrProxyServerReadErr() {
    proxyServerReadErrMu.Lock()
    proxyServerReadErr += 1
    proxyServerReadErrMu.Unlock()
}

func numProxyServerReadErr() (uint64) {
    return proxyServerReadErr
}

func incrProxyServerWriteErr() {
    proxyServerWriteErrMu.Lock()
    proxyServerWriteErr += 1
    proxyServerWriteErrMu.Unlock()
}

func numProxyServerWriteErr() (uint64) {
    return proxyServerWriteErr
}

func incrDirectServerReadErr() {
    directServerReadErrMu.Lock()
    directServerReadErr += 1
    directServerReadErrMu.Unlock()
}

func numDirectServerReadErr() (uint64) {
    return directServerReadErr
}

func incrDirectServerWriteErr() {
    directServerWriteErrMu.Lock()
    directServerWriteErr += 1
    directServerWriteErrMu.Unlock()
}

func numDirectServerWriteErr()(uint64) {
    return directServerWriteErr
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


