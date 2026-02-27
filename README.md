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

For cross-compilation to Windows on Linux, install the MinGW-w64 compiler:

```bash
sudo apt-get install gcc-mingw-w64
```

Then build:

```bash
CGO_ENABLED=1 GOOS=windows go build
```
