package main

import (
    "net"
    "testing"
)

//func TestNilCopy(t *testing.T) {
//    var dstname string = "destination"
//    var srcname string = "source"
//    copy(nil, nil, dstname, srcname)
//}

func TestNilClientToGetOriginalDst(t *testing.T) {
    getOriginalDst(nil)
}

func TestNilClientToHandleConnection(t *testing.T) {
    handleConnection(nil)
}

func TestNilClientToHandleDirectConnection(t *testing.T) {
    var ipv4 string = "1.2.3.4"
    var port uint16 = 8999
    
    // set up 
    gDirects = "1.2.3.4"
    dirFuncs := buildDirectors(gDirects)
    director = getDirector(dirFuncs)
    
    handleDirectConnection(nil, ipv4, port)
}

func TestNilClientToHandleProxyConnection(t *testing.T) {
    var ipv4 string = "2.3.4.5"
    var port uint16 = 8999
    handleProxyConnection(nil, ipv4, port)     
}

// when a &net.TCPConn{} is created, the underlying fd is set to nil. 
// make sure we can handle this situation without a panic (it has occurred before)
func TestEmptyFdToGetOriginalDst(t *testing.T) {
    var c1 *net.TCPConn
    c1 = &net.TCPConn{}
    getOriginalDst(c1)
}

func TestEmptyFdToHandleConnection(t *testing.T) {
    var c1 *net.TCPConn
    c1 = &net.TCPConn{}
    handleConnection(c1)
}

func TestEmptyFdToHandleDirectConnection(t *testing.T) {
    var ipv4 string = "1.2.3.4"
    var port uint16 = 8999

    // set up 
    gDirects = "1.2.3.4"
    dirFuncs := buildDirectors(gDirects)
    director = getDirector(dirFuncs)

    var c1 *net.TCPConn
    c1 = &net.TCPConn{}
    handleDirectConnection(c1, ipv4, port)
}

func TestEmptyFdToHandleProxyConnection(t *testing.T) {
    var ipv4 string = "2.3.4.5"
    var port uint16 = 8999
    var c1 *net.TCPConn
    c1 = &net.TCPConn{}
    handleProxyConnection(c1, ipv4, port)
}


