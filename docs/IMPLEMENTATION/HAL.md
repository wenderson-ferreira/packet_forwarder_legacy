# HAL interface implementation

The objective of this packet forwarder is to provide a lightweight implementation of the LoRaWAN specifications, adaptable to the different Hardware Abstraction Layers provided by Semtech and other actors. The logic behind the LoRaWAN protocol is thus loosely coupled to the specific concentrators interfaces. This means that it is possible for contributors to add compatibility to **new HALs**. For the moment, two HALs are available:

+ `halv1`, that interfaces with the classic SX1301 concentrator HAL. **This is the default HAL.**

+ `dummy`, that simulates an interaction with a concentrator. This HAL is to be reserved for testing purposes.

To add an interface with a HAL, you need to implement, in the `wrapper` package, all the methods that are called by the rest of the packet forwarder. You can refer to the `*_dummy.go` files, that contain the code for the dummy HAL, for this.

The classic process to add a new HAL is to add new files, in the `wrapper` package, that will **only build when the HAL identifier is passed as a build tag**. In Go, to specify this, you need to add a `// +build <tag>` at the beginning of the file. For example, for a new HAL called `devHAL`, this is what `gps_devhal.go` would look like:

```go
// +build devHAL

package wrapper

func LoRaGPSEnable(TTYPath string) error {
    return nil
}

// [...]
```

Once the development is over and you have implemented all `wrapper`'s functions for this new HAL, you can test the new HAL by building the packet forwarder by passing `HAL_CHOICE=<HAL identifier>` as environment variable:

```bash
$ export HAL_CHOICE=devHAL
$ make build
```
