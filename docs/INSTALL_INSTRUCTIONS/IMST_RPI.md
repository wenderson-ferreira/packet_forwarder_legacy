# Install the TTN Packet Forwarder on a Raspberry Pi with an IMST ic880a board

To follow this manual, you must have a Raspberry Pi with an IMST ic880a board, connected through SPI.

## Download and run

1. Download the [Raspberry Pi + IMST build](https://ttnreleases.blob.core.windows.net/packet-forwarder/master/imst-rpi-pktfwd.tar.gz) of the packet forwarder.

2. Configure the packet forwarder:

```bash
$ <packet-forwarder-binary> configure
[...]
  INFO New configuration file saved             ConfigFilePath=/root/.pktfwd.yml
```

3. Run the packet forwarder:

```bash
$ <packet-forwarder-binary> start
```

### Permanent installation with systemd

If you want a permanent installation of the packet forwarder on your Raspberry Pi, with `systemd` managing the packet forwarder on the background, we provide a basic systemd installation script, `install-systemd.sh`.

1. Select the build, and copy it in a permanent location - such as `/usr/bin`.

2. Create a configuration file in a permanent location, such as in a `/usr/config` directory:

```bash
$ touch /usr/config/ttn-pkt-fwd.yml
```

3. Set up this configuration file:

```bash
$ <packet-forwarder-binary> configure /usr/config/ttn-pkt-fwd.yml
```

4. Use the `install-systemd.sh` script, with the binary as a first argument and the config file as a second argument:

```bash
$ ./install-systemd.sh <packet-forwarder-binary-path> <configuration-file-path>
./install-systemd.sh: Installation of the systemd service complete.
```

5. Reload the systemd daemon, and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable ttn-pkt-fwd
sudo systemctl start ttn-pkt-fwd
```

## <a name="build"></a>Build

If want to contribute to the development of the packet forwarder, you might want to build the TTN Packet Forwarder. You will need to use a Linux environment to run the toolchain necessary for the build.

### Getting the toolchain

If you want to build the packet forwarder for a Raspberry Pi, you will need a **Raspberry Pi cross-compiler**. On some Linux distributions, such as Ubuntu, a toolchain is available as a package: `sudo apt install gcc-arm-linux-gnueabi -y`.

### Building the binary

Make sure you have [installed](https://golang.org/dl/) and [configured](https://golang.org/doc/code.html#GOPATH) your Go environment.

Follow these commands:

```bash
$ make dev-deps
$ make deps
$ GOOS=linux GOARCH=arm GOARM=7 CC=gcc-arm-linux-gnueabi make build
```

The binary will then be available in the `release/` folder.
