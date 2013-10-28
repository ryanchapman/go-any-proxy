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

func TestNilClient(t *testing.T) {
    var ipv4 string = "1.2.3.4"
    var port uint16 = 8999
//    handleConnection(nil)
//    handleDirectConnection(nil, ipv4, port)     
//    handleProxyConnection(nil, ipv4, port)     
    // create a conn, close it, then pass to handler funcs to make sure they can handle closed conn's appropriately
//    addr, err := net.ResolveTCPAddr("tcp", "www.google.com:80")
//    if err != nil {
//        t.Fatalf("Could not resolve www.google.com")
//    }
    var c1 *net.TCPConn
    c1 = nil
//    c1, err = net.DialTCP("tcp", nil, addr)
//    if err != nil {
//        t.Fatalf("Could not connect to www.google.com on port 80")
//    }
//    handleConnection(c1)
//    handleDirectConnection(c1, ipv4, port)
    handleProxyConnection(c1, ipv4, port)
}
