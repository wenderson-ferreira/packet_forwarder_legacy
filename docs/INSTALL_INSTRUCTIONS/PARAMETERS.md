# Build configuration

In addition to the hardware-specific parameters you can pass at build to enable or specify certain features, this document details the different parameters that can be used.

## HAL choice

For the moment, only two HALs are available. To switch HALs, pass the identifier of this HAL to `HAL_CHOICE`:

+ `halv1`, that interfaces with the classic SX1301 concentrator HAL. **This is the default value.**

+ `dummy`, that simulates an interaction with a concentrator. This HAL is to be reserved for testing purposes, on testing network environments.

To learn more about implementing an interface with another HAL, please consult the [implementation reference](../IMPLEMENTATION/HAL.md).
