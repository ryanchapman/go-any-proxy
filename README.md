[![Build Status](https://travis-ci.org/ryanchapman/go-any-proxy.png)](https://travis-ci.org/ryanchapman/go-any-proxy)

# Any Proxy

go-any-proxy is a server that can transparently proxy any tcp connection through an upstream proxy server.  This type
of setup is common in corporate environments.  It is written in golang and has been load tested with 10,000 concurrent
connections successfully on a Vyatta running a 64-bit kernel.

## Travis-CI

Build status can be found at http://travis-ci.org/ryanchapman/go-any-proxy

## More info

For more info, see http://blog.rchapman.org/post/47406142744/transparently-proxying-http-and-https-connections

## Authentication

You can add basic authentication parameters if needed, like this:

`any_proxy -l :3140 -p "MyLogin:Password25@proxy.corporate.com:8080"`

## Installation

```
$ git clone https://github.com/ryanchapman/go-any-proxy.git
$ cd go-any-proxy
$ ./make.bash
```

You'll end up with a binary `any_proxy`

## Experimental Mac OS X support
Fredrik Skogbreg has written the support for Mac OS X, but it is considered experimental until a load and performance
test is completed.  To build the mac version, after cloning this repo with `git clone https://github.com/ryanchapman/go-any-proxy.git`, 
change to the mac branch with `git checkout mac`, then make with `./make.bash`.  You'll need to configure some firewall
rules in Mac OS X firewall, see issue #16 (https://github.com/ryanchapman/go-any-proxy/pull/16) for instructions.


-Ryan A. Chapman<br>
 Sun Nov  2 16:39:24 MST 2014
