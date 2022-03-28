#!/bin/bash
go generate
go build -ldflags "-s -w" -trimpath
#chmod +x ./edm

