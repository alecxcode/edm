cd internal\config
go generate
cd ..\..
go generate
go build -ldflags "-H=windowsgui -s -w" -trimpath
