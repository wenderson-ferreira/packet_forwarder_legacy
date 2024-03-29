// Copyright © 2017 The Things Network. Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package wrapper

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/api/gateway"
)

const LengthPayload = 256 // length of the payload in bytes

// Packet describes the packets manipulated by the gateway
type Packet struct {
	Freq       uint32               // central frequency of the IF chain (in Hz)
	IFChain    uint8                // by which IF chain was packet received
	Status     uint8                // status of the received Packet
	CountUS    uint32               // internal concentrator counter for timestamping, 1 microsecond resolution
	Time       int64                // GPS-determined time
	Gps        *gateway.GPSMetadata // GPS Metadata
	RFChain    uint8                // by which RF chain was packet received
	Modulation uint8                // modulation used by the packet
	Bandwidth  uint8                // modulation bandwidth (LoRa only)
	Datarate   uint32               // RX datarate of the packet (SF for LoRa)
	Coderate   uint8                // error-correcting code of the packet (LoRa only)
	RSSI       float32              // average packet RSSI in dB
	SNR        float32              // average packet SNR, in dB (LoRa only)
	MinSNR     float32              // minimum packet SNR, in dB (LoRa only)
	MaxSNR     float32              // maximum packet SNR, in dB (LoRa only)
	CRC        uint16               // CRC that was received in the payload
	Size       uint32               // Payload size in bytes
	Payload    []byte               // Buffer containing the payload, not yet base64-encoded
}

type GPSCoordinates struct {
	Altitude  float64
	Latitude  float64
	Longitude float64
}

func (p Packet) DatarateString() (string, error) {
	if val, ok := datarateString[p.Datarate]; ok {
		return val, nil
	}
	return "", fmt.Errorf("LoRa packet with unknown datarate code %v", p.Datarate)
}

func (p Packet) BandwidthString() (string, error) {
	if val, ok := bandwidthString[p.Bandwidth]; ok {
		return val, nil
	}
	return "", fmt.Errorf("LoRa packet with unknown bandwidth code %v", p.Bandwidth)
}

func (p Packet) CoderateString() (string, error) {
	if val, ok := coderateString[p.Coderate]; ok {
		return val, nil
	}
	return "", fmt.Errorf("LoRa packet with unknown coderate code %v", p.Coderate)
}
