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


// Test if direct connections are working
// Should catch issue #11 if it occurs again 
// (shared memory issue related to the -d cmd line option)
func TestDirectConnectionFlags(t *testing.T) {
    // Test with the equivalent of a single IP address in the -d arg: -d 1.2.3.4
    gDirects = "1.2.3.4"
    dirFuncs := buildDirectors(gDirects)
    director = getDirector(dirFuncs)
    
    ipv4 := "1.2.3.4"
    wentDirect,_ := director(ipv4)
    if wentDirect == false {
        t.Errorf("The IP address %s should have been sent direct, but instead was proxied", ipv4)
    }

    // now make sure an address that should be proxied still works
    ipv4 = "4.5.6.7"
    wentDirect,_ = director(ipv4)
    if wentDirect == true {
        t.Errorf("The IP address %s should have been sent to an upstream proxy, but instead was sent directly", ipv4)
    }


    // Test with the equivalent of a multiple IP addresses in the -d arg: -d 1.2.3.4,2.3.4.5
    gDirects = "1.2.3.4,2.3.4.5"
    dirFuncs = buildDirectors(gDirects)
    director = getDirector(dirFuncs)
    
    addrsToTest := []string{"1.2.3.4", "2.3.4.5"}
    for _,ipv4 = range addrsToTest {
        wentDirect,_ := director(ipv4)
        if wentDirect == false {
            t.Errorf("The IP address %s should have been sent direct, but instead was proxied", ipv4)
        }
    }

    // now make sure an address that should be proxied still works
    ipv4 = "4.5.6.7"
    wentDirect,_ = director(ipv4)
    if wentDirect == true {
        t.Errorf("The IP address %s should have been sent to an upstream proxy, but instead was sent directly", ipv4)
    }


    // Test with the equivalent of multiple IP address specs in the -d arg: -d 1.2.3.0/24,2.3.4.0/25,4.4.4.4"
    gDirects = "1.2.3.0/24,2.3.4.0/25,4.4.4.4"
    dirFuncs = buildDirectors(gDirects)
    director = getDirector(dirFuncs)
    
    addrsToTest = []string{"1.2.3.4", "1.2.3.254", "2.3.4.5", "4.4.4.4"}
    for _,ipv4 = range addrsToTest {
        wentDirect,_ := director(ipv4)
        if wentDirect == false {
            t.Errorf("The IP address %s should have been sent direct, but instead was proxied", ipv4)
        }
    }

    // now make sure an address that should be proxied still works
    addrsToTest = []string{"4.5.6.7", "2.3.4.254"}
    for _,ipv4 = range addrsToTest {
        wentDirect,_ = director(ipv4)
        if wentDirect == true {
            t.Errorf("The IP address %s should have been sent to an upstream proxy, but instead was sent directly", ipv4)
        }
    }
}



