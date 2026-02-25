# sona-ui-client

## usage

port is sona's random port number

```
./sona-ui-client --port 34725 --forever
```

## linux

```
go build
```



## macos

```
CGO_ENABLED=1 GOOS=darwin go build
```



## windows

```
export GOOS=windows; export GOARCH=amd64; export CGO_ENABLED=1; export CXX=x86_64-w64-mingw32-g++; export CC=x86_64-w64-mingw32-gcc ; go build
```
