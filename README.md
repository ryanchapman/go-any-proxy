[![Build Status](https://travis-ci.org/ryanchapman/go-any-proxy.png)](https://travis-ci.org/ryanchapman/go-any-proxy)

# Any Proxy

go-any-proxy is a server that can transparently proxy any tcp connection through an upstream proxy server.  This type
of setup is common in corporate environments.  It is written in golang and has been load tested with 10,000 concurrent
connections successfully on a Vyatta running a 64-bit kernel.

## Travis-CI

Build status can be found at http://travis-ci.org/ryanchapman/go-any-proxy

## More info

For more info, see http://blog.rchapman.org/post/47406142744/transparently-proxying-http-and-https-connections

## Install Info 
You may need to run `go get github.com/zdannar/flogger` for library dependancies.

-Ryan A. Chapman
 Sat Jun  1 19:17:25 MDT 2013
