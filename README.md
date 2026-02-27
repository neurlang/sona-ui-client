# sona-ui-client

![L](sona.gif?raw=true)

**Features:**

- Start recording on click
- End recording on click/keypress
- Transcribe using local whisper model (to IPA or any Whisper-supported Language)
- Automatically copy transcribed text to clipboard

## usage

First install sona from [sona repo](https://github.com/thewh1teagle/sona).
Run sona with:

```
./sona serve model.bin
```
It will print the port:

```
{"commit":"dev","port":41911,"status":"ready","version":"dev"}
2026/02/27 20:37:24 listening on 0.0.0.0:41911
```

Then run this tool with the same port:

```
./sona-ui-client --port 41911
```

# Installation
## linux

```
go build
```

### linux - ubuntu - hotkey start the app and transcribe once

1. top right corner
2. settings
3. keyboard
4. keyboard shortcuts
5. custom shortcuts
6. add new
7. **name:** `sona` **command:** `/path/to/sona-ui-client --once` **shortcut:** set any key you like for sona

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
export GOOS=windows; export GOARCH=amd64; export CGO_ENABLED=1; export CXX=x86_64-w64-mingw32-g++; export CC=x86_64-w64-mingw32-gcc ; go build
```
