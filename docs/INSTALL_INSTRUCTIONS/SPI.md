# Install the TTN Packet Forwarder on a SPI environment

+ [Build procedure](#build)
+ [Cross-compilation](#crosscompilation)
+ [SPI configuration](#spi)
+ [GPS configuration](#gps)

## <a name="build"></a>Build procedure

Make sure you have [installed](https://golang.org/dl/) and [configured](https://golang.org/doc/code.html#GOPATH) Go environment 1.8.3 or higher.

This procedure describes how to build the packet forwarder for a machine that can interact with a concentrator using SPI. If you build the packet forwarder on the machine itself, the SPI configuration will be dynamically determined. Otherwise, see the [SPI configuration section](#spi) to see how to specify the SPI configuration.

```bash
$ go get -u github.com/TheThingsNetwork/packet_forwarder
$ cd $GOPATH/src/github.com/TheThingsNetwork/packet_forwarder
$ make dev-deps
$ make deps
$ make build
```

The build will then be available in the `release/` folder.

### <a name="crosscompilation"></a>Cross-compilation

If the gateway you wish to build the packet forwarder for doesn't support compilation, you will need to specify the target platform through environment variables:

* `GOOS`: OS of the target machine, following one of the values [supported by the Go toolchain](https://github.com/golang/go/blob/master/src/go/build/syslist.go). Example: `GOARCH=linux`.
* `GOARCH`: Architecture of the target machine, following one of the values [supported by the Go toolchain](https://github.com/golang/go/blob/master/src/go/build/syslist.go). Example: `GOARCH=amd64`, `GOARCH=arm`.
* `GOARM`: If building the packet forwarder for an `arm` architecture, the [ARM architecture to support](https://github.com/golang/go/wiki/GoArm). Example: `GOARM=5`, `GOARM=7`.
* `CC`: the `cc` compiler collection to use. Example: `CC=arm-none-linux-gnueabi-gcc`, `CC=arm-mlinux-gnuneabi-gcc`.
* `CROSS_COMPILE`: the prefix of the compiler collections to use. Example: `CROSS_COMPILE=arm-none-linux-gnueabi-`, `CROSS_COMPILE=arm-mlinux-gnuneabi-`.

These environment variables need to be set **just after the `make dev-deps`** step of the build process.

*Note: in several cases, toolchain setup scripts will set some of those variables for you, such as with the Multitech toolchain.*

## <a name="spi"></a>SPI configuration

If the build machine is different from the target machine, the SPI configuration can't be determined during the build process. You will then have to specify the configuration as a parameter.

* If the target machine is a `kerlink`, `imst_rpi`, `linklabs_blowfish_rpi` or `lorank` machine, the SPI configuration for those devices is already part of the HAL - you can build for those by specifying `PLATFORM=kerlink`, `PLATFORM=lorank`, and such. You can see `lora_gateway/libloragw/library.cfg`, once the dependencies have been installed.

* Otherwise, you can specify the parameters of the SPI interface with those parameters:

    * `SPI_SPEED` (default: `8000000`)
    * `SPIDEV` (default: the first `/dev/spidev*` file found on the build machine)
    * `SPI_CS_CHANGE` (default: `0`)
    * `VID` (default: `0x0403`)
    * `PID` (default: `0x6014`)

## <a name="gps"></a>GPS configuration

If the gateway you are running the packet forwarder has a GPS available, you can enable it by passing during the build the `GPS_PATH` environment variable. This `GPS_PATH` should point to the TTY path of the GPS. For example, on the Kerlink build, `GPS_PATH=/dev/nmea`.
