Flogger
=======

###SUMMARY

Flogger provides a simple wrapper around go's base logging module to provide
logging levels.  Flogger provides the following features.

* Logging levels capable of overloading output in desired language.
* Allow redirection of stdout and sterr to capture panics and other messages. 
* Other packages can import and use flogger without having to define a specific
  logger. 
* If a logfile is not configured, output will log to stdout.

###Example

```go
package main

import (
    log "flogger"
)

const (
    LOGFILE = "/tmp/example.log"
)


func main() {
    if err := log.OpenFile(LOGFILE, log.FLOG_APPEND, 0644); err != nil {
        log.Fatalf("Unable to open log file : %s", err)
    }
    log.SetLevel(log.INFO)

    log.RedirectStreams()

    log.Debugf("My name is %s", "mud")
    log.Infof("My name is %s", "Chuck Norris")
    log.Panic(" OH NO!... Don't panic!")
}
```

The output of /tmp/example.log

```
2013/10/13 17:33:38 logtest.go:21: : INFO : My name is Chuck Norris
2013/10/13 17:33:38 logtest.go:22: : PANIC :  OH NO!... Don't panic!
panic: : PANIC :  OH NO!... Don't panic!

goroutine 1 [running]:
github.com/ryanchapman/go-any-proxy/flogger.(*Flogger).flog(0xc20004b100, 0x5, 0x7f418ef8af78, 0x1, 0x1, ...)
        /home/zdannar/go_include/src/github.com/ryanchapman/go-any-proxy/flogger/flogger.go:71 +0x1e1
github.com/ryanchapman/go-any-proxy/flogger.(*Flogger).Panic(0xc20004b100, 0x7f418ef8af78, 0x1, 0x1)
        /home/zdannar/go_include/src/github.com/ryanchapman/go-any-proxy/flogger/flogger.go:136 +0x4c
github.com/ryanchapman/go-any-proxy/flogger.Panic(0x7f418ef8af78, 0x1, 0x1)
        /home/zdannar/go_include/src/github.com/ryanchapman/go-any-proxy/flogger/base.go:94 +0x46
main.main()
        /home/zdannar/gitroot/qa-site-tank/go_include/logtest.go:22 +0x30d

goroutine 2 [runnable]:
```
