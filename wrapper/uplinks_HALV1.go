// +build halv1

package wrapper

// #cgo CFLAGS: -I${SRCDIR}/../lora_gateway/libloragw/inc
// #cgo LDFLAGS: -lm ${SRCDIR}/../lora_gateway/libloragw/libloragw.a
// #include "config.h"
// #include "loragw_hal.h"
// #include "loragw_gps.h"
import "C"
import "errors"
import "time"

import "github.com/TheThingsNetwork/ttn/api/gateway"

const NbMaxPackets = 8
const nbRadios = C.LGW_RF_CHAIN_NB

const StatusCRCOK = uint8(C.STAT_CRC_OK)
const StatusCRCBAD = uint8(C.STAT_CRC_BAD)
const StatusNOCRC = uint8(C.STAT_NO_CRC)

const ModulationLoRa = uint8(C.MOD_LORA)
const ModulationFSK = uint8(C.MOD_FSK)

var datarateString = map[uint32]string{
	uint32(C.DR_LORA_SF7):  "SF7",
	uint32(C.DR_LORA_SF8):  "SF8",
	uint32(C.DR_LORA_SF9):  "SF9",
	uint32(C.DR_LORA_SF10): "SF10",
	uint32(C.DR_LORA_SF11): "SF11",
	uint32(C.DR_LORA_SF12): "SF12",
}

var bandwidthString = map[uint8]string{
	uint8(C.BW_125KHZ): "BW125",
	uint8(C.BW_250KHZ): "BW250",
	uint8(C.BW_500KHZ): "BW500",
}

var coderateString = map[uint8]string{
	uint8(C.CR_LORA_4_5): "4/5",
	uint8(C.CR_LORA_4_6): "4/6",
	uint8(C.CR_LORA_4_7): "4/7",
	uint8(C.CR_LORA_4_8): "4/8",
	0:                    "OFF",
}

// gpsReference is used to pass the GPS reference when building packets
type gpsReference struct {
	valid              bool
	validTimeReference bool
	timeReference      C.struct_tref
	locationReference  GPSCoordinates
}

func packetsFromCPackets(cPackets [8]C.struct_lgw_pkt_rx_s, nbPackets int) []Packet {
	var packetReference gpsReference
	if gpsActive() {
		// Using one global gpsReference avoids having one mutex lock per packet
		packetReference.valid = true
		packetReference.validTimeReference = checkGPSTimeReference()
		gpsTimeReferenceMutex.Lock()
		packetReference.timeReference = gpsTimeReference
		gpsTimeReferenceMutex.Unlock()
		coordinatesMutex.Lock()
		packetReference.locationReference = coordinates
		coordinatesMutex.Unlock()
	}

	var packets = make([]Packet, nbPackets)
	for i := 0; i < nbPackets && i < 8; i++ {
		packets[i] = packetFromCPacket(cPackets[i], packetReference)
	}
	return packets
}

func packetFromCPacket(cPacket C.struct_lgw_pkt_rx_s, currentReference gpsReference) Packet {
	// When using packetFromCPacket, it is assumed that accessing gpsTimeReferenceMutex
	// is safe => Use gpsTimeReferenceMutex before calling packetFromCPacket /before/
	// using this function
	var p = Packet{
		Freq:       uint32(cPacket.freq_hz),
		IFChain:    uint8(cPacket.if_chain),
		Status:     uint8(cPacket.status),
		CountUS:    uint32(cPacket.count_us),
		RFChain:    uint8(cPacket.rf_chain),
		Modulation: uint8(cPacket.modulation),
		Bandwidth:  uint8(cPacket.bandwidth),
		Datarate:   uint32(cPacket.datarate),
		Coderate:   uint8(cPacket.coderate),
		RSSI:       float32(cPacket.rssi),
		SNR:        float32(cPacket.snr),
		MinSNR:     float32(cPacket.snr_min),
		MaxSNR:     float32(cPacket.snr_max),
		CRC:        uint16(cPacket.crc),
		Size:       uint32(cPacket.size),
	}

	p.Payload = make([]byte, p.Size)
	var i uint32
	for i = 0; i < p.Size; i++ {
		p.Payload[i] = byte(cPacket.payload[i])
	}

	if currentReference.valid {
		p.Gps = &gateway.GPSMetadata{
			Latitude:  float32(currentReference.locationReference.Latitude),
			Longitude: float32(currentReference.locationReference.Longitude),
			Altitude:  int32(currentReference.locationReference.Altitude),
		}

		var pktUtcTime C.struct_timespec
		if currentReference.validTimeReference && C.lgw_cnt2utc(currentReference.timeReference, cPacket.count_us, &pktUtcTime) == C.LGW_GPS_SUCCESS {
			// conversion successful
			p.Time = time.Unix(int64(pktUtcTime.tv_sec), int64(pktUtcTime.tv_nsec)).UnixNano()
			p.Gps.Time = p.Time
		}
	}
	return p
}

func Receive() ([]Packet, error) {
	var packets [NbMaxPackets]C.struct_lgw_pkt_rx_s
	concentratorMutex.Lock()
	nbPackets := C.lgw_receive(NbMaxPackets, &packets[0])
	concentratorMutex.Unlock()
	if nbPackets == C.LGW_HAL_ERROR {
		return nil, errors.New("Failed packet fetch from the concentrator")
	}
	return packetsFromCPackets(packets, int(nbPackets)), nil
}
