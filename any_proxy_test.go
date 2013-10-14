package main

import (
    "testing"
)

func TestNilCopy(t *testing.T) {
    var dstname string = "destination"
    var srcname string = "source"
    copy(nil, nil, dstname, srcname)
}

func TestNilClient(t *testing.T) {
    var ipv4 string = "1.2.3.4"
    var port uint16 = 8999
    handleConnection(nil)
    handleDirectConnection(nil, ipv4, port)     
    handleProxyConnection(nil, ipv4, port)     
}
