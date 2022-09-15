#!/bin/bash
cd internal/config
go generate
cd ../..
go generate
go build -ldflags "-s -w" -trimpath
#chmod +x ./edm

