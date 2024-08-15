set GOOS=windows
: set GOARCH=386
: set AR=i686-w64-mingw32-ar
: set CC=i686-w64-mingw32-gcc
: set CXX=i686-w64-mingw32-g++
set GOARCH=amd64
set AR=x86_64-w64-mingw32-ar
set CC=x86_64-w64-mingw32-gcc
set CXX=x86_64-w64-mingw32-g++
: set CGO_ENABLED=1

@set PATH=%CYGWIN64DIR%\bin;%PATH%

set TARGET=tunproxy

@del build\%TARGET%.exe

: go fmt %TARGET%
: go fmt %TARGET%/config
: go fmt %TARGET%/service
: go fmt %TARGET%/startup
: go fmt %TARGET%/toolkit

: golangci-lint run . service toolkit

: protoc --goV2_out=. service/pb/*.proto

: windres -o %TARGET%.syso %TARGET%.rc

go build -o build/%TARGET%.exe -ldflags "-w -s" %TARGET%
: go build -o build/%TARGET%.exe -race -ldflags "-H windowsgui -w -s" %TARGET%

@pause