# Building toolchain images

This document describes how to build the toolchain Docker images, to build them for development purposes. Indeed, some of the toolchains (such as the Kerlink IoT Station toolchain) are privately-licensed - we thus cannot publicly release them. The Dockerfiles mentioned in this document are located in the `scripts/toolchains` folder.

Once the images are built on your personal machine, you can set up your own CI pipeline, using the [GitLab CI](.gitlab-ci.yml) configuration file in the repo.

* [Kerlink IoT Station](#klk-iot-station)
* [Multitech Conduit](#multitech)

## <a name="klk-iot-station"></a>Kerlink IoT Station

To build the `registry.gitlab.com/thethingsnetwork/packet_forwarder/klk-toolchain` image, you will need access to Kerlink's Wirnet Station wiki.

1. On Kerlink's Wirnet Station Wiki, click on *Resources*, scroll down to *Tools*, then download the toolchain you need, depending on the firmware of your Station.

2. In most cases, the archive will hold a `arm-2011.03-wirgrid` folder. Copy this folder in `packet-forwarder/scripts/toolchains`.

3. Build the image: `docker build . -t registry.gitlab.com/thethingsnetwork/packet_forwarder/klk-toolchain -f Dockerfile.kerlink-iot-station`. The content of the toolchain will be copied to form the image.

## <a name="multitech"></a>Multitech Conduit mLinux

The Multitech Conduit image, tagged `registry.gitlab.com/thethingsnetwork/packet_forwarder/multitech-toolchain`, does not need access to any private resource, and can be built solely with its Dockerfile:

```bash
$ docker build . -t registry.gitlab.com/thethingsnetwork/packet_forwarder/multitech-toolchain -f Dockerfile.multitech
```

*Note: depending on the Docker storage driver on which the image is built, the SDK extraction can fail. In that case, you might want to change the storage driver to `aufs`.*
