# Downlinks implementation

According to TTN specifications, gateways don't decide on which reception window should be sent a downlink. The TTN back-end decides, according to its own logic, when a downlink should be sent. The objectives of this implementation of this packet forwarder, in terms of downlink reception and transmission, are:

* Transmitting every downlink packet within a reasonable timing to make sure that the reception window requirements of every one of them is met ;

* Making an optimal use of the concentrator buffer, to be able to send as much downlinks as possible.

Most concentrators only have a single downlink buffer - which means that the packet forwarder has to handle the logic of transmitting packets at the right moment to lose a minimum amount of packets.

+ [`sendingTimeMargin` values](#values)

## Downlink emission process

To handle the distribution of a downlink, a packet forwarder must transmit to the concentrator, with the packet, the **internal clock time** at which the **concentrator should emit it**. This value is called `ExpectedSendingTimestamp`. The internal clock from the concentrator is initiated at 0 when the concentrator is started - during the `lgw_start()` HAL function call that starts the concentrator. This initialisation moment is called `ConcentratorBootTime`. The `lgw_start()` call lasting a few seconds, saving the time reference from the moment the HAL function was called is not precise enough.

The method we use to find `ConcentratorBootTime` is through the first uplink. With every uplink, a value `count_us` is transmitted to the packet forwarder, that contains the **value of the concentrator's internal clock** at the uplink reception in the concentrator. In the manager (`pktfwd/manager.go`), during the first uplink reception, the calculation to find `ConcentratorBootTime` is made from this `count_us` value from the current time.

It is important to note that because of the uplink polling rate, `count_us` only allows us to find `ConcentratorBootTime` within 100μs. When the packet forwarder starts, it polls for uplinks every 100μs, to have a higher degree of precision for the `ConcentratorBootTime` value calculation. When the first uplink has been received, the polling frequency is diminished to every 5ms, to avoid performance issues.

* When a downlink is received, the packet forwarder schedules it in an internal queue system to be handled **100ms before `ExpectedSendingTimestamp`**. This means that the packet forwarder has then 100ms to perform its last computations on the downlink packet and to transmit it. This 100ms margin value is called `sendingTimeMargin`.
	
	* The value of `sendingTimeMargin` has a consequence of the gateway's downlink debit rate. Considering we need 100ms to transmit a downlink from the gateway's internal memory to emission, it means that we can only reasonably transmit 600 downlinks per minute - and that is making the assumption that receive windows won't overlap.

	* Having a 100ms `sendingTimeMargin` allows the packet forwarder to have a comfortable margin in case of performance issues on the system, or in case of transmission issues. For systems connected to a concentrator via USB, it usually takes 10ms to perform the last computations and to transmit the packet to the concentrator. However, one improvement to the packet forwarder would be setting `sendingTimeMargin` as a build or run parameter, to make use of the higher transmission speeds on SPI-connected devices.

*Note:* The packet forwarder doesn't support GPS concentrators yet. GPS concentrators don't rely on an internal clock, and are able to transmit absolute timestamps for an uplink - meaning it is not necessary to know their internal clock value to transmit downlinks to such devices.

## <a name="values"></a>Specific `sendingTimeMargin` values

Depending on the hardware, we might change the value of `sendingTimeMargin` to adapt. A higher `sendingTimeMargin` value means more risk of having downlinks being deleted by the packet forwarder before sent, but can be necessary for the downlinks to be sent in time.

|Build|`sendingTimeMargin` value|
|---|---|
|Kerlink IoT Station|50ms|
|Multitech Conduit|100ms|
