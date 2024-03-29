# Install the TTN Packet Forwarder on a Kerlink IoT Station

*Note: for the moment, the TTN Packet Forwarder is not compatible with the Kerlink iBTS.*

Before installing the TTN Packet Forwarder, we recommend **updating the Station to the latest firmware available**.

+ [Download and test the TTN Packet Forwarder](#download-test)
+ [Install the TTN Packet Forwarder](#install)
+ [Build the TTN Packet Forwarder](#build)
+ [Troubleshooting](#troubleshooting)

## <a name="download-test"></a>Download and test the TTN Packet Forwarder

*Note: Before installing the new packet forwarder, make sure you removed any other packet forwarder installed on your Kerlink IoT Station. If you don't have any important files stored on the disk, the safest way to make sure of that is to update the Station to the latest firmware available, which will reset the file system in the process.*

1. Download the [Kerlink build](https://ttnreleases.blob.core.windows.net/packet-forwarder/master/kerlink-iot-station-pktfwd.tar.gz) of the packet forwarder.

2. In the folder, you will find several files: a `create-package.sh` script and a binary file, that we will call `packet-forwarder`.

The binary is sufficient for a testing use - if you wish to try the TTN packet forwarder, just copy the binary on the Station, and execute:

```bash
$ ./packet-forwarder configure
# Follow the instructions of the wizard
[...]
  INFO New configuration file saved             ConfigFilePath=/root/config.yml
$ ./packet-forwarder start
  INFO Packet Forwarder for LoRa Gateway        HALVersionInfo=Version: 4.0.0; Options: native;
[...]
  INFO Concentrator started, packets can now be received and sent
```

## <a name="install"></a>Install the TTN Packet Forwarder

This section covers permanent installation of the TTN Packet Forwarder on a Kerlink IoT Station.

### Packaging the TTN Packet Forwarder

Download the [Kerlink build](https://ttnreleases.blob.core.windows.net/packet-forwarder/master/kerlink-iot-station-pktfwd.tar.gz) of the packet forwarder. Execute the `create-package.sh` script with the binary inside as an argument:

```bash
$ ./create-package.sh packet-forwarder
[...]
# The script will ask you several questions to configure the packet forwarder.
./create-package.sh: Kerlink DOTA package complete.
```

A `kerlink-release-<id>` folder will appear in the folder you are in:

```bash
$ cd kerlink-release-<id> && tree
.
├── dota_ttn-pkt-fwd.tar.gz
├── INSTALL.md
└── produsb.sh
```

### Transfering and installing the TTN Packet Forwarder

You can consult the `INSTALL.md` to know how to install the package from here. The two options are **network transfer** and **USB stick transfer**. Depending on your configuration, choose the installation method that suits you best. Once the package has been transferred, the packet forwarder will be installed on your Kerlink IoT Station! You can monitor it from the [console](https://console.thethingsnetwork.org):

![Console demo](https://github.com/TheThingsNetwork/packet_forwarder/raw/master/docs/INSTALL_INSTRUCTIONS/console.gif)

## <a name="build"></a>Build the TTN Packet Forwarder for the Kerlink IoT Station

If you use a specific machine or want to contribute to the development of the packet forwarder, you might want to build the TTN Packet Forwarder. You might need to use a Linux environment to run the toolchain necessary for the build.

### Building the binary

To build the packet forwarder for the Kerlink IoT Station, you will need access to Kerlink's Wirnet Station wiki.

1. On Kerlink's Wirnet Station Wiki, click on *Resources*, scroll down to *Tools*, then download the toolchain you need, depending on the firmware of your Station.

2. In most cases, the archive will hold a `arm-2011.03-wirgrid` folder. Copy this folder in `/opt`:

```bash
$ mkdir -p /opt
$ mv arm-2011.03-wirgrid /opt
```

3. Execute the `scripts/build-kerlink.sh` script, indicating in argument the location of the toolchain:

```bash
$ ./scripts/build-kerlink.sh "/opt/arm-2011.03-wirgrid"
```

This script will build the Kerlink IoT Station binary of the packet forwarder.

### Building the DOTA file

This binary is sufficient for basic testing of the packet forwarder on a Kerlink IoT Station. However, for permanent installations, the packet forwarder is wrapped in a package called DOTA file. To create a DOTA file, use the `scripts/kerlink/create-kerlink-package.sh` script:

```bash
$ ./scripts/kerlink/create-kerlink-package.sh <packet-forwarder-binary-path>
[...]
create-kerlink-package.sh: Kerlink DOTA package complete. The package is available in kerlink-release/. Consult the INSTALL.md file to know how to install the package on your Kerlink IoT Station!
```

## <a name="troubleshooting"></a>Troubleshooting

#### I've deleted my gateway from the console and added a new one, how can I change the configuration of the gateway?

Connect remotely to the Kerlink IoT Station, and execute these commands:

```bash
$ cd /mnt/fsuser-1/ttn-packet-forwarder
$ ./ttn-pkt-fwd configure config.yml
# Following the instructions of the wizard
[...]
  INFO New configuration file saved             ConfigFilePath=config.yml
$ reboot
# Reboot the gateway to apply the changes
```

#### I'm getting a "Concentrator boot time computation error: Absurd uptime received by the concentrator" error when starting the packet forwarder.

The concentrator sometimes sends absurd uptime values to the packet forwarder, often because it hasn't been stopped properly. Restart the packet forwarder until this error disappears.
