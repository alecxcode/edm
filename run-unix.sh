#!/bin/zsh
cd internal/config
go generate
cd ../..
go generate
go build
#chmod +x ./edm
./edm --consolelog
