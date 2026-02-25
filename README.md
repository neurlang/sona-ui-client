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
CGO_ENABLED=1 GOOS=windows go build
```

but currently, needs fix:
```
# runtime/cgo
gcc: error: unrecognized command-line option ‘-mthreads’; did you mean ‘-pthread’?
```
