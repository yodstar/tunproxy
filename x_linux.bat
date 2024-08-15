set GOOS=linux
: set GOARCH=386
: set AR=i686-w64-mingw32-ar
: set CC=i686-w64-mingw32-gcc
: set CXX=i686-w64-mingw32-g++
: set GOARCH=amd64
: set AR=x86_64-w64-mingw32-ar
: set CC=x86_64-w64-mingw32-gcc
: set CXX=x86_64-w64-mingw32-g++
: set CGO_ENABLED=1

@set PATH=%CYGWIN64DIR%\bin;%PATH%

set TARGET=tunproxy

@del build\%TARGET%

: go fmt %TARGET%
: go fmt %TARGET%/config
: go fmt %TARGET%/service
: go fmt %TARGET%/startup
: go fmt %TARGET%/toolkit

: golangci-lint run . config controller model service startup toolkit

go build -o build/%TARGET% -ldflags "-w -s" %TARGET%

@pause