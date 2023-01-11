cd internal\config
go generate
cd ..\..\internal\core
go generate
cd ..\..\cmd\edm
go build -ldflags "-H=windowsgui -s -w" -trimpath
cd ..\..
move cmd\edm\edm.exe edm.exe
