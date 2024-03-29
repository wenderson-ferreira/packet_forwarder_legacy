# Install the TTN Packet Forwarder on a Multitech Conduit

*Note: if you're using an AEP model, you will need to configure the Conduit on the web interface before installing the Packet Forwarder. Consult [this guide](https://www.thethingsnetwork.org/docs/gateways/multitech/aep.html) to learn how to do this.*

## Download and install

*Note: Before installing the new packet forwarder, make sure you removed any other packet forwarder installed on your Multitech Conduit.*

1. Download the [Multitech Conduit package](https://ttnreleases.blob.core.windows.net/packet-forwarder/master/multitech-conduit-pktfwd.tar.gz) of the packet forwarder.

2. In the archive, you will find an `create-package.sh` file, a `multitech-installer.sh`, as well as the executable binary. Execute the `create-package.sh` file, with the binary as a first argument:

```bash
$ ./create-package.sh <packet-forwarder-binary>
[...]
# Following the instructions of the wizard
./create-package.sh: package available at ttn-pkt-fwd.ipk
```

3. Copy the package on the Multitech Conduit, using either a USB key or `scp` if you have an SSH connection to the Multitech Conduit. Install the package, configure the packet forwarder, then start it:

```bash
$ opkg install ttn-pkt-fwd.ipk
Installing ttn-pkt-fwd (2.0.0) to root...
Configuring ttn-pkt-fwd.
$ /etc/init.d/ttn-pkt-fwd configure
[...]
# Following the instructions of the wizard
  INFO New configuration file saved             ConfigFilePath=/var/config/ttn-pkt-fwd/config.yml
$ /etc/init.d/ttn-pkt-fwd start
Starting ttn-pkt-fwd: OK
```

## <a name="build"></a>Build the TTN Packet Forwarder for the Multitech Conduit

If you use a specific machine or want to contribute to the development of the packet forwarder, you might want to build the TTN Packet Forwarder. You might need to use a Linux environment to run the toolchain necessary for the build.

### Downloading the Multitech toolchain

To build the packet forwarder, you will need to download [Multitech's C toolchain](http://www.multitech.net/developer/software/mlinux/mlinux-software-development/mlinux-c-toolchain/). Download it, and install it by following the instructions on Multitech's website.

### Building the binary

1. Download the packet forwarder, along with its dependencies:

```bash
$ go get -u github.com/TheThingsNetwork/packet_forwarder
$ cd $GOPATH/src/github.com/TheThingsNetwork/packet_forwarder
$ make dev-deps
$ make deps
```

2. Enable the Multitech toolchain, then set those environment variables:

```bash
$ source /path/to/sdk/environment-setup-arm926ejste-mlinux-linux-gnueabi
# Usually /opt/mlinux/{version}/environment-setup-arm926ejste-mlinux-linux-gnueabi
$ export GOARM=5
$ export GOOS=linux
$ export GOARCH=arm
$ export CFG_SPI=ftdi
$ export PLATFORM=multitech
```

3. Build the binary:

```bash
$ make build
```

The binary will then be available in the `release/` folder.

### Building the package

To build the package, use the `scripts/multitech/create-package.sh` script:

```bash
$ ./scripts/multitech/create-package.sh release/<packet-forwarder-binary>
[...]
# Following the instructions of the wizard
./create-package.sh: package available at ttn-pkt-fwd.ipk
```

The package will then be available at the specified path.
