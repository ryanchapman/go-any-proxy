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
//
//
// Tested to 2000 connections/second.  If you turn off logging, you can get 10,000/sec. So logging needs
// to be changed to nonblocking one day.
//
// TODO:
// add num of connected clients to stats
// add ability to print details of each connected client (src,dst,proxy or direct addr) to stats
//
// Ryan A. Chapman, ryan@rchapman.org
// Sun Apr  7 21:04:34 MDT 2013
//


package main

import (
    "bufio"
    "errors"
    "flag"
    "fmt"
    "io"
    log "github.com/zdannar/flogger"
    "net"
    "os"
    "os/signal"
    "runtime"
    "runtime/pprof"
    "strconv"
    "strings"
    "syscall"
    "time"
    "encoding/base64"
)

const VERSION = "1.1"
const SO_ORIGINAL_DST = 80
const DEFAULTLOG = "/var/log/any_proxy.log"
const STATSFILE  = "/var/log/any_proxy.stats"


var gListenAddrPort string
var gProxyServerSpec string
var gDirects string
var gVerbosity int
var gProxyServers []string
var gAuthProxyServers = map[string] string { }
var gLogfile string
var gCpuProfile string
var gMemProfile string
var gClientRedirects int
var gReverseLookups int

type cacheEntry struct {
    hostname string
    expires time.Time
}
type reverseLookupCache struct {
  hostnames map[string]*cacheEntry
  keys []string
  next int
}
func NewReverseLookupCache() *reverseLookupCache {
    return &reverseLookupCache{
        hostnames: make(map[string]*cacheEntry),
        keys: make([]string,65536),
    }
}
func (c *reverseLookupCache) lookup(ipv4 string) string {
    hit := c.hostnames[ipv4]
    if hit != nil {
        if hit.expires.After(time.Now()) {
            log.Debugf("lookup(): CACHE_HIT")
            return hit.hostname
        } else {
            log.Debugf("lookup(): CACHE_EXPIRED")
            delete(c.hostnames, ipv4)
        }
    } else {
        log.Debugf("lookup(): CACHE_MISS")
    }
    return ""
}
func (c *reverseLookupCache) store(ipv4, hostname string) {
    delete(c.hostnames, c.keys[c.next])
    c.keys[c.next] = ipv4
    c.next = (c.next + 1) & 65535
    c.hostnames[ipv4] = &cacheEntry{hostname: hostname, expires: time.Now().Add(time.Hour)}
}
var gReverseLookupCache *reverseLookupCache

type directorFunc func(*net.IP) bool
var director func(*net.IP) (bool, int)

func init() {
    flag.Usage = func() {
        fmt.Fprintf(os.Stdout, "%s\n\n", versionString())
        fmt.Fprintf(os.Stdout, "usage: %s -l listenaddress -p proxies [-d directs] [-v=N] [-f file] [-c file] [-m file]\n", os.Args[0])
        fmt.Fprintf(os.Stdout, "       Proxies any tcp port transparently using Linux netfilter\n\n")
        fmt.Fprintf(os.Stdout, "Mandatory\n")
        fmt.Fprintf(os.Stdout, "  -l=ADDRPORT      Address and port to listen on (e.g., :3128 or 127.0.0.1:3128)\n")
        fmt.Fprintf(os.Stdout, "Optional\n")
        fmt.Fprintf(os.Stdout, "  -p=PROXIES       Address and ports of upstream proxy servers to use\n")
        fmt.Fprintf(os.Stdout, "                   Multiple address/ports can be specified by separating with commas\n")
        fmt.Fprintf(os.Stdout, "                   (e.g., 10.1.1.1:80,10.2.2.2:3128 would try to proxy requests to a\n")
        fmt.Fprintf(os.Stdout, "                    server listening on port 80 at 10.1.1.1 and if that failed, would\n")
        fmt.Fprintf(os.Stdout, "                    then try port 3128 at 10.2.2.2)\n")
        fmt.Fprintf(os.Stdout, "                   Note that requests are not load balanced. If a request fails to the\n")
        fmt.Fprintf(os.Stdout, "                   first proxy, then the second is tried and so on.\n\n")
        fmt.Fprintf(os.Stdout, "  -d=DIRECTS       List of IP addresses that the proxy should send to directly instead of\n")
        fmt.Fprintf(os.Stdout, "                   to the upstream proxies (e.g., -d 10.1.1.1,10.1.1.2)\n")
        fmt.Fprintf(os.Stdout, "  -r=1             Enable relaying of HTTP redirects from upstream to clients\n")
        fmt.Fprintf(os.Stdout, "  -R=1             Enable reverse lookups of destination IP address and use hostname in CONNECT\n")
        fmt.Fprintf(os.Stdout, "                   request instead of the numeric IP if available. A local DNS server could be\n")
        fmt.Fprintf(os.Stdout, "                   configured to provide a reverse lookup of the forward lookup responses seen.\n")
        fmt.Fprintf(os.Stdout, "  -v=1             Print debug information to logfile %s\n", DEFAULTLOG)
        fmt.Fprintf(os.Stdout, "  -f=FILE          Log file. If not specified, defaults to %s\n", DEFAULTLOG)
        fmt.Fprintf(os.Stdout, "  -c=FILE          Write a CPU profile to FILE. The pprof program, which is part of Golang's\n")
        fmt.Fprintf(os.Stdout, "                   standard pacakge, can be used to interpret the results. You can invoke pprof\n")
        fmt.Fprintf(os.Stdout, "                   with \"go tool pprof\"\n")
        fmt.Fprintf(os.Stdout, "  -m=FILE          Write a memory profile to FILE. This file can also be interpreted by golang's pprof\n\n")
        fmt.Fprintf(os.Stdout, "  -h               This usage message\n")
        fmt.Fprintf(os.Stdout, "any_proxy should be able to achieve 2000 connections/sec with logging on, 10k with logging off (-f=/dev/null).\n")
        fmt.Fprintf(os.Stdout, "Before starting any_proxy, be sure to change the number of available file handles to at least 65535\n")
        fmt.Fprintf(os.Stdout, "with \"ulimit -n 65535\"\n")
        fmt.Fprintf(os.Stdout, "Some other tunables that enable higher performance:\n")
        fmt.Fprintf(os.Stdout, "  net.core.netdev_max_backlog = 2048\n")
        fmt.Fprintf(os.Stdout, "  net.core.somaxconn = 1024\n")
        fmt.Fprintf(os.Stdout, "  net.core.rmem_default = 8388608\n")
        fmt.Fprintf(os.Stdout, "  net.core.rmem_max = 16777216\n")
        fmt.Fprintf(os.Stdout, "  net.core.wmem_max = 16777216\n")
        fmt.Fprintf(os.Stdout, "  net.ipv4.ip_local_port_range = 2000 65000\n")
        fmt.Fprintf(os.Stdout, "  net.ipv4.tcp_window_scaling = 1\n")
        fmt.Fprintf(os.Stdout, "  net.ipv4.tcp_max_syn_backlog = 3240000\n")
        fmt.Fprintf(os.Stdout, "  net.ipv4.tcp_max_tw_buckets = 1440000\n")
        fmt.Fprintf(os.Stdout, "  net.ipv4.tcp_mem = 50576 64768 98152\n")
        fmt.Fprintf(os.Stdout, "  net.ipv4.tcp_rmem = 4096 87380 16777216\n")
        fmt.Fprintf(os.Stdout, "  NOTE: if you see syn flood warnings in your logs, you need to adjust tcp_max_syn_backlog, tcp_synack_retries and tcp_abort_on_overflow\n");
        fmt.Fprintf(os.Stdout, "  net.ipv4.tcp_syncookies = 1\n")
        fmt.Fprintf(os.Stdout, "  net.ipv4.tcp_wmem = 4096 65536 16777216\n")
        fmt.Fprintf(os.Stdout, "  net.ipv4.tcp_congestion_control = cubic\n\n")
        fmt.Fprintf(os.Stdout, "To obtain statistics, send any_proxy signal SIGUSR1. Current stats will be printed to %v\n", STATSFILE)
        fmt.Fprintf(os.Stdout, "Report bugs to <ryan@rchapman.org>.\n") 
    }
    flag.StringVar(&gListenAddrPort,  "l", "", "Address and port to listen on")
    flag.StringVar(&gProxyServerSpec, "p", "", "Proxy servers to use, separated by commas. E.g. -p proxy1.tld.com:80,proxy2.tld.com:8080,proxy3.tld.com:80")
    flag.StringVar(&gDirects,         "d", "", "IP addresses to go direct")
    flag.StringVar(&gLogfile,         "f", "", "Log file")
    flag.StringVar(&gCpuProfile,      "c", "", "Write cpu profile to file")
    flag.StringVar(&gMemProfile,      "m", "", "Write mem profile to file")
    flag.IntVar(   &gVerbosity,       "v", 0,  "Control level of logging. v=1 results in debugging info printed to the log.\n")
    flag.IntVar(   &gClientRedirects, "r", 0,  "Should we relay HTTP redirects from upstream proxies? -r=1 if we should.\n")
    flag.IntVar(   &gReverseLookups,  "R", 0,  "Should we perform reverse lookups of destination IPs and use hostnames? -h=1 if we should.\n")

    dirFuncs := buildDirectors(gDirects)
    director = getDirector(dirFuncs)
}

func versionString() (v string) {
    buildNum := strings.ToUpper(strconv.FormatInt(BUILDTIMESTAMP, 36))
    buildDate := time.Unix(BUILDTIMESTAMP, 0).Format(time.UnixDate)
    v = fmt.Sprintf("any_proxy %s (build %v, %v by %v@%v)", VERSION, buildNum, buildDate, BUILDUSER, BUILDHOST)
    return
}

func buildDirectors(gDirects string) ([]directorFunc) {
    // Generates a list of directorFuncs that are have "cached" values within
    // the scope of the functions.  

    directorCidrs := strings.Split(gDirects, ",")
    directorFuncs := make([]directorFunc, len(directorCidrs))

    for idx,directorCidr := range directorCidrs {
        //dstring := director
        var dfunc directorFunc
        if strings.Contains(directorCidr, "/") {
            _, directorIpNet, err := net.ParseCIDR(directorCidr)
            if err != nil {
                panic(fmt.Sprintf("\nUnable to parse CIDR string : %s : %s\n", directorCidr, err))
            }
            dfunc = func(ptestip *net.IP) bool {
                testIp := *ptestip
                return directorIpNet.Contains(testIp)
            }
            directorFuncs[idx] = dfunc
        } else {
            var directorIp net.IP
            directorIp = net.ParseIP(directorCidr)
            dfunc = func(ptestip *net.IP) bool {
                var testIp net.IP
                testIp = *ptestip
                return testIp.Equal(directorIp)
            }
            directorFuncs[idx] = dfunc
        }

    }
    return directorFuncs
}

func getDirector(directors []directorFunc) func(*net.IP) (bool, int) {
    // getDirector:
    // Returns a function(directorFunc) that loops through internally held 
    // directors evaluating each for possible matches.
    // 
    // directorFunc: 
    // Loops through directors and returns the (true, idx) where the index is 
    // the sequential director that returned true. Else the function returns
    // (false, 0) if there are no directors to handle the ip.

    dFunc := func(ipaddr *net.IP) (bool, int) {
        for idx, dfunc := range directors {
            if dfunc(ipaddr) {
                return true, idx
            }
        }
        return false, 0
    }
    return dFunc
}

func setupProfiling() {
    // Make sure we have enough time to write profile's to disk, even if user presses Ctrl-C
    if gMemProfile == "" || gCpuProfile == "" {
        return
    }

    var profilef *os.File
    var err error
    if gMemProfile != "" {
        profilef, err = os.Create(gMemProfile)
        if err != nil {
            panic(err)
        }
    }

    if gCpuProfile != "" {
        f, err := os.Create(gCpuProfile)
        if err != nil {
            panic(err)
        }
        pprof.StartCPUProfile(f)
    }

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func() {
        for _ = range c {
            if gCpuProfile != "" {
                pprof.StopCPUProfile()
            }
            if gMemProfile != "" {
                pprof.WriteHeapProfile(profilef)
                profilef.Close()
            }
            time.Sleep(5000 * time.Millisecond)
            os.Exit(0)
        }
    }()
}

func setupLogging() {
    if gLogfile == "" {
        gLogfile = DEFAULTLOG
    }

    log.SetLevel(log.INFO)
    if gVerbosity != 0 {
        log.SetLevel(log.DEBUG)
    }

    if err := log.OpenFile(gLogfile, log.FLOG_APPEND, 0644); err != nil {
        log.Fatalf("Unable to open log file : %s", err)
    }
}

func main() {
    flag.Parse()
    if gListenAddrPort == "" {
        flag.Usage()
        os.Exit(1)
    }

    runtime.GOMAXPROCS(runtime.NumCPU() / 2)
    setupLogging()
    setupProfiling()
    setupStats()

    dirFuncs := buildDirectors(gDirects)
    director = getDirector(dirFuncs)

    if gReverseLookups == 1 {
        gReverseLookupCache = NewReverseLookupCache()
    }

    log.RedirectStreams()

    // if user gave us upstream proxies, check and see if they are alive
    if gProxyServerSpec != "" {
	checkProxies()
    }

    lnaddr, err := net.ResolveTCPAddr("tcp", gListenAddrPort)
    if err != nil {
        panic(err)
    }

    listener, err := net.ListenTCP("tcp", lnaddr)
    if err != nil {
        panic(err)
    }
    defer listener.Close()
    log.Infof("Listening for connections on %v\n", listener.Addr())

    for {
        conn, err := listener.AcceptTCP()
        if err != nil {
            log.Infof("Error accepting connection: %v\n", err)
            incrAcceptErrors()
            continue
        }
        incrAcceptSuccesses()
        go handleConnection(conn)
    }
}

func checkProxies() {
    gProxyServers = strings.Split(gProxyServerSpec, ",")
    // make sure proxies resolve and are listening on specified port
    for i, proxySpec := range gProxyServers {
	if strings.Contains(proxySpec, "@") {
	    var authSplit = strings.Split(proxySpec, "@")
	    var b64Auth = base64.StdEncoding.EncodeToString([]byte(authSplit[0]))
	    gAuthProxyServers[authSplit[1]] = b64Auth
	    proxySpec = authSplit[1]
	    gProxyServers[i] = proxySpec
	    log.Infof("Added authentication %v, %v\n", authSplit[0], b64Auth)
	}

        conn, err := dial(proxySpec)
        if err != nil {
            log.Infof("Test connection to %v: failed. Removing from proxy server list\n", proxySpec)
            a := gProxyServers[:i]
            b := gProxyServers[i+1:]
            gProxyServers = append(a, b...)
            continue
        }
        conn.Close()
        log.Infof("Added proxy server %v\n", proxySpec)
    }
    // do we have at least one proxy server?
    if len(gProxyServers) == 0 {
        msg := "None of the proxy servers specified are available. Exiting."
        log.Infof("%s\n", msg)
        fmt.Fprintf(os.Stderr, msg)
        os.Exit(1)
    }
}

func copy(dst io.ReadWriteCloser, src io.ReadWriteCloser, dstname string, srcname string) {
    if dst == nil {
        log.Debugf("copy(): oops, dst is nil!")
        return
    }
    if src == nil {
        log.Debugf("copy(): oops, src is nil!")
        return
    }
    _, err := io.Copy(dst, src)
    if err != nil {
        if operr, ok := err.(*net.OpError); ok {
            if srcname == "directserver" || srcname == "proxyserver" {
                log.Debugf("copy(): %s->%s: Op=%s, Net=%s, Addr=%v, Err=%v", srcname, dstname, operr.Op, operr.Net, operr.Addr, operr.Err)
            }
            if operr.Op == "read" {
                if srcname == "proxyserver" {
                    incrProxyServerReadErr()
                }
                if srcname == "directserver" {
                    incrDirectServerReadErr()
                }
            }
            if operr.Op == "write" {
                if srcname == "proxyserver" {
                    incrProxyServerWriteErr()
                }
                if srcname == "directserver" {
                    incrDirectServerWriteErr()
                }
            }
        }
    }
    dst.Close()
    src.Close()
}

func getOriginalDst(clientConn *net.TCPConn) (ipv4 string, port uint16, newTCPConn *net.TCPConn, err error) {
    if clientConn == nil {
        log.Debugf("copy(): oops, dst is nil!")
        err = errors.New("ERR: clientConn is nil")
        return
    }

    // test if the underlying fd is nil
    remoteAddr := clientConn.RemoteAddr()
    if remoteAddr == nil {
        log.Debugf("getOriginalDst(): oops, clientConn.fd is nil!")
        err = errors.New("ERR: clientConn.fd is nil")
        return
    }

    srcipport := fmt.Sprintf("%v", clientConn.RemoteAddr())

    newTCPConn = nil
    // net.TCPConn.File() will cause the receiver's (clientConn) socket to be placed in blocking mode.
    // The workaround is to take the File returned by .File(), do getsockopt() to get the original 
    // destination, then create a new *net.TCPConn by calling net.Conn.FileConn().  The new TCPConn
    // will be in non-blocking mode.  What a pain.
    clientConnFile, err := clientConn.File()
    if err != nil {
        log.Infof("GETORIGINALDST|%v->?->FAILEDTOBEDETERMINED|ERR: could not get a copy of the client connection's file object", srcipport)
        return
    } else {
        clientConn.Close()
    }

    // Get original destination
    // this is the only syscall in the Golang libs that I can find that returns 16 bytes
    // Example result: &{Multiaddr:[2 0 31 144 206 190 36 45 0 0 0 0 0 0 0 0] Interface:0}
    // port starts at the 3rd byte and is 2 bytes long (31 144 = port 8080)
    // IPv4 address starts at the 5th byte, 4 bytes long (206 190 36 45)
    addr, err :=  syscall.GetsockoptIPv6Mreq(int(clientConnFile.Fd()), syscall.IPPROTO_IP, SO_ORIGINAL_DST)
    log.Debugf("getOriginalDst(): SO_ORIGINAL_DST=%+v\n", addr)
    if err != nil {
        log.Infof("GETORIGINALDST|%v->?->FAILEDTOBEDETERMINED|ERR: getsocketopt(SO_ORIGINAL_DST) failed: %v", srcipport, err)
        return
    }
    newConn, err := net.FileConn(clientConnFile)
    if err != nil {
        log.Infof("GETORIGINALDST|%v->?->%v|ERR: could not create a FileConn fron clientConnFile=%+v: %v", srcipport, addr, clientConnFile, err)
        return
    }
    if _, ok := newConn.(*net.TCPConn); ok {
        newTCPConn = newConn.(*net.TCPConn)
        clientConnFile.Close()
    } else {
        errmsg := fmt.Sprintf("ERR: newConn is not a *net.TCPConn, instead it is: %T (%v)", newConn, newConn)
        log.Infof("GETORIGINALDST|%v->?->%v|%s", srcipport, addr, errmsg)
        err = errors.New(errmsg)
        return
    }

    ipv4 = itod(uint(addr.Multiaddr[4])) + "." + 
           itod(uint(addr.Multiaddr[5])) + "." + 
           itod(uint(addr.Multiaddr[6])) + "." + 
           itod(uint(addr.Multiaddr[7]))
    port = uint16(addr.Multiaddr[2]) << 8 + uint16(addr.Multiaddr[3])

    return
}

func dial(spec string) (*net.TCPConn, error) {
    host, port, err := net.SplitHostPort(spec)
    if err != nil {
        log.Infof("dial(): ERR: could not extract host and port from spec %v: %v", spec, err)
        return nil, err
    }
    remoteAddr, err := net.ResolveIPAddr("ip", host)
    if err != nil {
        log.Infof("dial(): ERR: could not resolve %v: %v", host, err)
        return nil, err
    }
    portInt, err := strconv.Atoi(port)
    if err != nil {
        log.Infof("dial(): ERR: could not convert network port from string \"%s\" to integer: %v", port, err)
        return nil, err
    }
    remoteAddrAndPort := &net.TCPAddr{IP: remoteAddr.IP, Port: portInt}
    var localAddr *net.TCPAddr
    localAddr = nil
    conn, err := net.DialTCP("tcp", localAddr, remoteAddrAndPort)
    if err != nil {
        log.Infof("dial(): ERR: could not connect to %v:%v: %v", remoteAddrAndPort.IP, remoteAddrAndPort.Port, err)
    }
    return conn, err
}

func handleDirectConnection(clientConn *net.TCPConn, ipv4 string, port uint16) {
    // TODO: remove
    log.Debugf("Enter handleDirectConnection: clientConn=%+v (%T)\n", clientConn, clientConn)

    if clientConn == nil {
        log.Debugf("handleDirectConnection(): oops, clientConn is nil!")
        return
    }

    // test if the underlying fd is nil
    remoteAddr := clientConn.RemoteAddr()
    if remoteAddr == nil {
        log.Debugf("handleDirectConnection(): oops, clientConn.fd is nil!")
        return
    }

    ipport := fmt.Sprintf("%s:%d", ipv4, port)
    directConn, err := dial(ipport)
    if err != nil {
        log.Infof("DIRECT|%v->%v|Could not connect, giving up: %v", clientConn.RemoteAddr(), directConn.RemoteAddr(), err)
        return
    }
    log.Debugf("DIRECT|%v->%v|Connected to remote end", clientConn.RemoteAddr(), directConn.RemoteAddr())
    incrDirectConnections()
    go copy(clientConn, directConn, "client", "directserver")
    go copy(directConn, clientConn, "directserver", "client")
}

func handleProxyConnection(clientConn *net.TCPConn, ipv4 string, port uint16) {
    var proxyConn net.Conn
    var err error
    var success bool = false
    var host string
    var headerXFF string = ""

    // TODO: remove
    log.Debugf("Enter handleProxyConnection: clientConn=%+v (%T)\n", clientConn, clientConn)

    if clientConn == nil {
        log.Debugf("handleProxyConnection(): oops, clientConn is nil!")
        return
    }

    // test if the underlying fd is nil
    remoteAddr := clientConn.RemoteAddr()
    if remoteAddr == nil {
        log.Debugf("handleProxyConnect(): oops, clientConn.fd is nil!")
        err = errors.New("ERR: clientConn.fd is nil")
        return
    }

    host, _, err = net.SplitHostPort(remoteAddr.String())
    if err == nil {
        headerXFF = fmt.Sprintf("X-Forwarded-For: %s\r\n", host)
    }

    if gReverseLookups == 1 {
        hostname := gReverseLookupCache.lookup(ipv4)
        if hostname != "" {
            ipv4 = hostname
        } else {
            names, err := net.LookupAddr(ipv4)
            if err == nil && len(names) > 0 {
                gReverseLookupCache.store(ipv4,names[0])
                ipv4 = names[0]
            }
        }
    }

    for _, proxySpec := range gProxyServers {
        proxyConn, err = dial(proxySpec)
        if err != nil {
            log.Debugf("PROXY|%v->%v->%s:%d|Trying next proxy.", clientConn.RemoteAddr(), proxySpec, ipv4, port)
            continue
        }
        log.Debugf("PROXY|%v->%v->%s:%d|Connected to proxy\n", clientConn.RemoteAddr(), proxyConn.RemoteAddr(), ipv4, port)
	var authString = ""
	if val, auth := gAuthProxyServers[proxySpec]; auth {
	  authString = fmt.Sprintf("\r\nProxy-Authorization: Basic %s", val)
	}
        connectString := fmt.Sprintf("CONNECT %s:%d HTTP/1.0%s\r\n%s\r\n", ipv4, port, authString, headerXFF)
        log.Debugf("PROXY|%v->%v->%s:%d|Sending to proxy: %s\n", clientConn.RemoteAddr(), proxyConn.RemoteAddr(), ipv4, port, strconv.Quote(connectString))
        fmt.Fprintf(proxyConn, connectString)
        status, err := bufio.NewReader(proxyConn).ReadString('\n')
        log.Debugf("PROXY|%v->%v->%s:%d|Received from proxy: %s", clientConn.RemoteAddr(), proxyConn.RemoteAddr(), ipv4, port, strconv.Quote(status))
        if err != nil {
            log.Infof("PROXY|%v->%v->%s:%d|ERR: Could not find response to CONNECT: err=%v. Trying next proxy", clientConn.RemoteAddr(), proxyConn.RemoteAddr(), ipv4, port, err)
            incrProxyNoConnectResponses()
            continue
        }
        if strings.Contains(status, "400") { // bad request
            log.Debugf("PROXY|%v->%v->%s:%d|Status from proxy=400 (Bad Request)", clientConn.RemoteAddr(), proxyConn.RemoteAddr(), ipv4, port)
            log.Debugf("%v: Response from proxy=400", proxySpec)
            incrProxy400Responses()
            copy(clientConn, proxyConn, "client", "proxyserver")
            return
        }
        if strings.Contains(status, "301") || strings.Contains(status, "302") && gClientRedirects == 1 {
            log.Debugf("PROXY|%v->%v->%s:%d|Status from proxy=%s (Redirect), relaying response to client", clientConn.RemoteAddr(), proxyConn.RemoteAddr(), ipv4, port, strconv.Quote(status))
            incrProxy300Responses()
            fmt.Fprintf(clientConn, status)
            copy(clientConn, proxyConn, "client", "proxyserver")
            return
        }
        if strings.Contains(status, "200") == false {
            log.Infof("PROXY|%v->%v->%s:%d|ERR: Proxy response to CONNECT was: %s. Trying next proxy.\n", clientConn.RemoteAddr(), proxyConn.RemoteAddr(), ipv4, port, strconv.Quote(status))
            incrProxyNon200Responses()
            continue
        } else {
            incrProxy200Responses()
        }
        log.Debugf("PROXY|%v->%v->%s:%d|Proxied connection", clientConn.RemoteAddr(), proxyConn.RemoteAddr(), ipv4, port)
        success = true
        break
    }
    if proxyConn == nil {
        log.Debugf("handleProxyConnection(): oops, proxyConn is nil!")
        return
    }
    if success == false {
        log.Infof("PROXY|%v->UNAVAILABLE->%s:%d|ERR: Tried all proxies, but could not establish connection. Giving up.\n", clientConn.RemoteAddr(), ipv4, port)
        fmt.Fprintf(clientConn, "HTTP/1.0 503 Service Unavailable\r\nServer: go-any-proxy\r\nX-AnyProxy-Error: ERR_NO_PROXIES\r\n\r\n")
        clientConn.Close()
        return
    } 
    incrProxiedConnections()
    go copy(clientConn, proxyConn, "client", "proxyserver")
    go copy(proxyConn, clientConn, "proxyserver", "client")
}

func handleConnection(clientConn *net.TCPConn) {
    if clientConn == nil {
        log.Debugf("handleConnection(): oops, clientConn is nil")
        return
    }

    // test if the underlying fd is nil
    remoteAddr := clientConn.RemoteAddr()
    if remoteAddr == nil {
        log.Debugf("handleConnection(): oops, clientConn.fd is nil!")
        return
    }

    ipv4, port, clientConn, err := getOriginalDst(clientConn)
    if err != nil {
        log.Infof("handleConnection(): can not handle this connection, error occurred in getting original destination ip address/port: %+v\n", err)
        return
    }
    // If no upstream proxies were provided on the command line, assume all traffic should be sent directly
    if gProxyServerSpec == "" {
            handleDirectConnection(clientConn, ipv4, port)
            return
    } 
    // Evaluate for direct connection
    ip := net.ParseIP(ipv4)
    if ok,_ := director(&ip); ok {
            handleDirectConnection(clientConn, ipv4, port)
            return
    }
    handleProxyConnection(clientConn, ipv4, port)
}

// from pkg/net/parse.go
// Convert i to decimal string.
func itod(i uint) string {
        if i == 0 {
                return "0"
        }

        // Assemble decimal in reverse order.
        var b [32]byte
        bp := len(b)
        for ; i > 0; i /= 10 {
                bp--
                b[bp] = byte(i%10) + '0'
        }

        return string(b[bp:])
}
