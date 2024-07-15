[![any-proxy Build and Test](https://github.com/ryanchapman/go-any-proxy/actions/workflows/any-proxy.yml/badge.svg)](https://github.com/ryanchapman/go-any-proxy/actions/workflows/any-proxy.yml)

# Any Proxy

go-any-proxy is a server that can transparently proxy any tcp connection through an upstream proxy server.  This type
of setup is common in corporate environments.  It is written in golang and has been load tested with 10,000 concurrent
connections successfully on a Vyatta running a 64-bit kernel.

## More info

For more info, see http://blog.rchapman.org/post/47406142744/transparently-proxying-http-and-https-connections

## Maintenance

This project is actively maintained.  As of this writing (Jult 2024), I haven't had many bugs submitted in a few
years, which is why you don't see much for code changing.  But be assured that I am watching the project and will
address any bugs that come in.

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
