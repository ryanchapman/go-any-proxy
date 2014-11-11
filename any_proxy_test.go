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
    
    ipv4 := net.ParseIP("1.2.3.4")
    wentDirect,_ := director(&ipv4)
    if wentDirect == false {
        t.Errorf("The IP address %s should have been sent direct, but instead was proxied", ipv4)
    }

    // now make sure an address that should be proxied still works
    ipv4 = net.ParseIP("4.5.6.7")
    wentDirect,_ = director(&ipv4)
    if wentDirect == true {
        t.Errorf("The IP address %s should have been sent to an upstream proxy, but instead was sent directly", ipv4)
    }


    // Test with the equivalent of a multiple IP addresses in the -d arg: -d 1.2.3.4,2.3.4.5
    gDirects = "1.2.3.4,2.3.4.5"
    dirFuncs = buildDirectors(gDirects)
    director = getDirector(dirFuncs)
    
    addrsToTest := []net.IP{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")}
    for _,ipv4 = range addrsToTest {
        wentDirect,_ := director(&ipv4)
        if wentDirect == false {
            t.Errorf("The IP address %s should have been sent direct, but instead was proxied", ipv4)
        }
    }

    // now make sure an address that should be proxied still works
    ipv4 = net.ParseIP("4.5.6.7")
    wentDirect,_ = director(&ipv4)
    if wentDirect == true {
        t.Errorf("The IP address %s should have been sent to an upstream proxy, but instead was sent directly", ipv4)
    }


    // Test with the equivalent of multiple IP address specs in the -d arg: -d 1.2.3.0/24,2.3.4.0/25,4.4.4.4"
    gDirects = "1.2.3.0/24,2.3.4.0/25,4.4.4.4"
    dirFuncs = buildDirectors(gDirects)
    director = getDirector(dirFuncs)
    
    addrsToTest = []net.IP{net.ParseIP("1.2.3.4"), net.ParseIP("1.2.3.254"), net.ParseIP("2.3.4.5"), net.ParseIP("4.4.4.4")}
    for _,ipv4 = range addrsToTest {
        wentDirect,_ := director(&ipv4)
        if wentDirect == false {
            t.Errorf("The IP address %s should have been sent direct, but instead was proxied", ipv4)
        }
    }

    // now make sure an address that should be proxied still works
    addrsToTest = []net.IP{net.ParseIP("4.5.6.7"), net.ParseIP("2.3.4.254")}
    for _,ipv4 = range addrsToTest {
        wentDirect,_ = director(&ipv4)
        if wentDirect == true {
            t.Errorf("The IP address %s should have been sent to an upstream proxy, but instead was sent directly", ipv4)
        }
    }
}

// benchmark when we have 1 direct. The address we are testing against is one that
// will not match any directs, just to make sure we search through all directs
func BenchmarkDirector1(b *testing.B) {
    gDirects = "1.2.3.0/24"
    dirFuncs := buildDirectors(gDirects)
    director := getDirector(dirFuncs)

    ipv4 := net.ParseIP("255.255.255.0")
    for n := 0; n < b.N; n++ {
        director(&ipv4)
    }
}

func BenchmarkDirector2(b *testing.B) {
    gDirects = "1.0.0.0/24,2.0.0.0/24"
    dirFuncs := buildDirectors(gDirects)
    director := getDirector(dirFuncs)

    ipv4 := net.ParseIP("255.255.255.0")
    for n := 0; n < b.N; n++ {
        director(&ipv4)
    }
}

func BenchmarkDirector3(b *testing.B) {
    gDirects = "1.0.0.0/24,2.0.0.0/24,3.0.0.0/24"
    dirFuncs := buildDirectors(gDirects)
    director := getDirector(dirFuncs)

    ipv4 := net.ParseIP("255.255.255.0")
    for n := 0; n < b.N; n++ {
        director(&ipv4)
    }
}

func BenchmarkDirector4(b *testing.B) {
    gDirects = "1.0.0.0/24,2.0.0.0/24,3.0.0.0/24,4.0.0.0/24"
    dirFuncs := buildDirectors(gDirects)
    director := getDirector(dirFuncs)

    ipv4 := net.ParseIP("255.255.255.0")
    for n := 0; n < b.N; n++ {
        director(&ipv4)
    }
}

func BenchmarkDirector5(b *testing.B) {
    gDirects = "1.0.0.0/24,2.0.0.0/24,3.0.0.0/24,4.0.0.0/24,5.0.0.0/24"
    dirFuncs := buildDirectors(gDirects)
    director := getDirector(dirFuncs)

    ipv4 := net.ParseIP("255.255.255.0")
    for n := 0; n < b.N; n++ {
        director(&ipv4)
    }
}

func BenchmarkDirector10(b *testing.B) {
    gDirects  = "1.0.0.0/24,2.0.0.0/24,3.0.0.0/24,4.0.0.0/24,5.0.0.0/24,"
    gDirects += "6.0.0.0/24,7.0.0.0/24,8.0.0.0/24,9.0.0.0/24,10.0.0.0/24"
    dirFuncs := buildDirectors(gDirects)
    director := getDirector(dirFuncs)

    ipv4 := net.ParseIP("255.255.255.0")
    for n := 0; n < b.N; n++ {
        director(&ipv4)
    }
}

func BenchmarkDirector100(b *testing.B) {
    gDirects  = "1.0.0.0/24,2.0.0.0/24,3.0.0.0/24,4.0.0.0/24,5.0.0.0/24,"
    gDirects += "6.0.0.0/24,7.0.0.0/24,8.0.0.0/24,9.0.0.0/24,10.0.0.0/24,"
    gDirects += "11.0.0.0/24,12.0.0.0/24,13.0.0.0/24,14.0.0.0/24,15.0.0.0/24,"
    gDirects += "16.0.0.0/24,17.0.0.0/24,18.0.0.0/24,19.0.0.0/24,20.0.0.0/24,"
    gDirects += "21.0.0.0/24,22.0.0.0/24,23.0.0.0/24,24.0.0.0/24,25.0.0.0/24,"
    gDirects += "26.0.0.0/24,27.0.0.0/24,28.0.0.0/24,29.0.0.0/24,30.0.0.0/24,"
    gDirects += "31.0.0.0/24,32.0.0.0/24,33.0.0.0/24,34.0.0.0/24,35.0.0.0/24,"
    gDirects += "36.0.0.0/24,37.0.0.0/24,38.0.0.0/24,39.0.0.0/24,40.0.0.0/24,"
    gDirects += "41.0.0.0/24,42.0.0.0/24,43.0.0.0/24,44.0.0.0/24,45.0.0.0/24,"
    gDirects += "46.0.0.0/24,47.0.0.0/24,48.0.0.0/24,49.0.0.0/24,50.0.0.0/24,"
    gDirects += "51.0.0.0/24,52.0.0.0/24,53.0.0.0/24,54.0.0.0/24,55.0.0.0/24,"
    gDirects += "56.0.0.0/24,57.0.0.0/24,58.0.0.0/24,59.0.0.0/24,60.0.0.0/24,"
    gDirects += "61.0.0.0/24,62.0.0.0/24,63.0.0.0/24,64.0.0.0/24,65.0.0.0/24,"
    gDirects += "66.0.0.0/24,67.0.0.0/24,68.0.0.0/24,69.0.0.0/24,70.0.0.0/24,"
    gDirects += "71.0.0.0/24,72.0.0.0/24,73.0.0.0/24,74.0.0.0/24,75.0.0.0/24,"
    gDirects += "76.0.0.0/24,77.0.0.0/24,78.0.0.0/24,79.0.0.0/24,80.0.0.0/24,"
    gDirects += "81.0.0.0/24,82.0.0.0/24,83.0.0.0/24,84.0.0.0/24,85.0.0.0/24,"
    gDirects += "86.0.0.0/24,87.0.0.0/24,88.0.0.0/24,89.0.0.0/24,90.0.0.0/24,"
    gDirects += "91.0.0.0/24,92.0.0.0/24,93.0.0.0/24,94.0.0.0/24,95.0.0.0/24,"
    gDirects += "96.0.0.0/24,97.0.0.0/24,98.0.0.0/24,99.0.0.0/24,100.0.0.0/24"
    dirFuncs := buildDirectors(gDirects)
    director := getDirector(dirFuncs)

    ipv4 := net.ParseIP("255.255.255.0")
    for n := 0; n < b.N; n++ {
        director(&ipv4)
    }
}
