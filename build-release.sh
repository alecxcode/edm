#!/bin/sh
cd internal/config
go generate
cd ../../internal/core
go generate
cd ../../cmd/edm
go build -ldflags "-s -w" -trimpath
cd ../..
mv cmd/edm/edm ./edm
#chmod +x ./edm

