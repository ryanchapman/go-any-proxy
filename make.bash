#!/bin/bash

# Required if you are using a 32-bit vyatta. Otherwise, comment it out
GOOS=linux
#GOARCH=386

go build any_proxy.go
