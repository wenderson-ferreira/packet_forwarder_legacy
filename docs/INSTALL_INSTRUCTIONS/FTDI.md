# Install the TTN Packet Forwarder on a FTDI environment

*Note: Support of FTDI is experimental and not guaranteed, as Semtech has dropped FTDI support for the Hardware Abstraction Layer. macOS builds should only be used for development purposes.*

+ [Build procedure](#build)
+ [macOS troubleshooting](#macos)

## <a name="build"></a>Build procedure

Building the packet forwarder on a FTDI environment, with a USB connection with the concentrator, requires the `libmpsse` library.

### Install `libmpsse`

1. `brew install libftdi` or `apt install libftdi`
2. `wget https://storage.googleapis.com/google-code-archive-downloads/v2/code.google.com/libmpsse/libmpsse-1.3.tar.gz`
3. `tar -xvzf libmpsse-1.3.tar.gz`
4. `cd libmpsse-1.3/src`
5. `./configure --disable-python && make && sudo make install`

### Download and build the packet forwarder

Make sure you have [installed](https://golang.org/dl/) and [configured](https://golang.org/doc/code.html#GOPATH) your Go environment.

```bash
$ go get -u github.com/TheThingsNetwork/packet_forwarder
$ cd $GOPATH/src/github.com/TheThingsNetwork/packet_forwarder
$ make dev-deps
$ make deps
# If you are using Linux:
$ CFG_SPI=ftdi PLATFORM=imst_rpi make build
# If you are using macOS:
$ CFG_SPI=mac PLATFORM=imst_rpi make build
```

The build will then be available in the `release/` folder.

## <a name="macos"></a>macOS troubleshooting

On a macOS environment, you will need, at every reboot, to unload the native Apple Driver for FTDI devices: `sudo kextunload -b com.apple.driver.AppleUSBFTDI`. If you are unsure of the name of the driver, you can look for it with the command `kextstat | grep FTDI`.
