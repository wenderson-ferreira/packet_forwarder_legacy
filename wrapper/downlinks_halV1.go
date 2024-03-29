// +build halv1

package wrapper

// #cgo CFLAGS: -I${SRCDIR}/../lora_gateway/libloragw/inc
// #cgo LDFLAGS: -lm ${SRCDIR}/../lora_gateway/libloragw/libloragw.a
// #include "config.h"
// #include "loragw_hal.h"
// #include "loragw_gps.h"
import "C"

import (
	"errors"
	"fmt"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/packet_forwarder/util"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/api/router"
)

const (
	stdFSKPreamble  = 4
	stdLoRaPreamble = 8
	fieldInfo       = 0
	crcPoly16       = uint16(0x1021)
	crcInitVal16    = uint16(0xFFFF)
)

var coderateValueMap = map[string]C.uint8_t{
	"4/5": C.CR_LORA_4_5,
	"4/6": C.CR_LORA_4_6,
	"2/3": C.CR_LORA_4_6,
	"4/7": C.CR_LORA_4_7,
	"4/8": C.CR_LORA_4_8,
	"1/2": C.CR_LORA_4_8,
}

func coderateValue(i string) (C.uint8_t, error) {
	if val, ok := coderateValueMap[i]; ok {
		return val, nil
	}
	return 0, errors.New("TX packet with unknown coderate")
}

func bandwidthValue(i uint32) (C.uint8_t, error) {
	if val, ok := loraChannelBandwidths[i]; ok {
		return val, nil
	}
	return 0, errors.New("TX packet with unknown bandwidth")
}

func sfValue(i uint32) (C.uint32_t, error) {
	if val, ok := loraChannelSpreadingFactors[i]; ok {
		return val, nil
	}
	return 0, errors.New("TX packet with unknown spreading factor")
}

func getLoRaDatarate(datarateStr string) (C.uint32_t, C.uint8_t, error) {
	var sf, bw uint32
	var err error
	var bandwidth C.uint8_t
	var spreadingFactor C.uint32_t
	nb, err := fmt.Sscanf(datarateStr, "SF%dBW%d", &sf, &bw)
	if err != nil {
		return 0, 0, err
	}

	if nb != 2 {
		return 0, 0, errors.New("Couldn't parse LoRa datarate for the downlink message - aborting this TX packet")
	}

	spreadingFactor, err = sfValue(sf)
	if err != nil {
		return 0, 0, errors.New("Couldn't read LoRa datarate for the downlink message (unknown Spreading Factor value)")
	}

	bandwidth, err = bandwidthValue(bw * 1000)
	if err != nil {
		return 0, 0, errors.New("Couldn't read LoRa datarate for the downlink message (unknown Bandwidth value)")
	}

	return spreadingFactor, bandwidth, nil
}

func setupLoRaDownlink(txPacket *C.struct_lgw_pkt_tx_s, downlink router.DownlinkMessage) error {
	txPacket.modulation = C.MOD_LORA
	var err error
	txPacket.datarate, txPacket.bandwidth, err = getLoRaDatarate(downlink.GetProtocolConfiguration().GetLorawan().GetDataRate())
	if err != nil {
		return err
	}
	txPacket.coderate, err = coderateValue(downlink.GetProtocolConfiguration().GetLorawan().GetCodingRate())
	if err != nil {
		return err
	}
	txPacket.invert_pol = C.bool(downlink.GetGatewayConfiguration().GetPolarizationInversion())
	txPacket.preamble = C.uint16_t(stdLoRaPreamble)
	return nil
}

func setupFSKDownlink(txPacket *C.struct_lgw_pkt_tx_s, downlink router.DownlinkMessage) {
	txPacket.modulation = C.MOD_FSK
	txPacket.preamble = C.uint16_t(stdFSKPreamble)
	txPacket.datarate = C.uint32_t(downlink.GetProtocolConfiguration().GetLorawan().GetBitRate())
	txPacket.f_dev = C.uint8_t(downlink.GetGatewayConfiguration().GetFrequencyDeviation() / 1000) /* gRPC value in Hz, txpkt.f_dev in kHz */
}

func checkRFPower(cconf util.SX1301Conf, downlink router.DownlinkMessage) error {
	for _, val := range cconf.GetTXLuts() {
		if val.RfPower == int8(downlink.GetGatewayConfiguration().GetPower()) {
			return nil
		}
	}
	return errors.New("Unsupported RF Power for TX")
}

func setupDownlinkModulation(downlink router.DownlinkMessage, txPacket *C.struct_lgw_pkt_tx_s) error {
	if downlink.GetProtocolConfiguration().GetLorawan().GetModulation() == lorawan.Modulation_LORA {
		return setupLoRaDownlink(txPacket, downlink)
	} else if downlink.GetProtocolConfiguration().GetLorawan().GetModulation() == lorawan.Modulation_FSK {
		setupFSKDownlink(txPacket, downlink)
		return nil
	}
	return errors.New("Modulation neither LoRa nor FSK")
}

func insertPayload(downlink router.DownlinkMessage, txPacket *C.struct_lgw_pkt_tx_s) error {
	payload := downlink.GetPayload()
	if len(payload) > 256 {
		return errors.New("Payload too big to transmit")
	}
	txPacket.size = C.uint16_t(len(payload))
	for i := 0; i < len(payload); i++ {
		txPacket.payload[i] = C.uint8_t(payload[i])
	}
	return nil
}

func SendDownlink(downlink *router.DownlinkMessage, conf util.Config, ctx log.Interface) error {
	var txPacket = C.struct_lgw_pkt_tx_s{
		freq_hz:   C.uint32_t(downlink.GetGatewayConfiguration().GetFrequency()),
		rf_chain:  C.uint8_t(downlink.GetGatewayConfiguration().GetRfChain()),
		no_crc:    C.bool(false),
		no_header: C.bool(false),
		payload:   [256]C.uint8_t{},
		tx_mode:   C.TIMESTAMPED,
		count_us:  C.uint32_t(downlink.GetGatewayConfiguration().GetTimestamp()),
	}

	// Inserting payload
	if err := insertPayload(*downlink, &txPacket); err != nil {
		ctx.WithError(err).Warn("Failure parsing and wrapping the current TX packet - aborting transmission")
		return err
	}

	// Antenna gain
	if antennaGain := conf.Concentrator.AntennaGain; antennaGain != nil {
		txPacket.rf_power = C.int8_t(downlink.GetGatewayConfiguration().GetPower() - int32(*antennaGain))
	} else {
		txPacket.rf_power = C.int8_t(downlink.GetGatewayConfiguration().GetPower())
	}

	// LoRa/FSK parameters
	if err := setupDownlinkModulation(*downlink, &txPacket); err != nil {
		ctx.WithError(err).Warn("Failure parsing and wrapping the current TX packet during the parameter verification - aborting transmission")
		return err
	}

	// Checking RFPower
	if err := checkRFPower(conf.Concentrator, *downlink); err != nil {
		ctx.WithError(err).Warn("Failure parsing and wrapping the current TX packet during the RFPower check - aborting transmission")
		return err
	}

	return sendDownlinkConcentrator(txPacket, ctx)
}

func sendDownlinkConcentrator(txPacket C.struct_lgw_pkt_tx_s, ctx log.Interface) error {
	for {
		var txStatus C.uint8_t
		concentratorMutex.Lock()
		var result = C.lgw_status(C.TX_STATUS, &txStatus)
		concentratorMutex.Unlock()
		if result == C.LGW_HAL_ERROR {
			ctx.Warn("Couldn't get concentrator status")
		} else if txStatus == C.TX_EMITTING {
			// XX: Should we stop emission (like in the legacy packet forwarder) or retry?
			// If we retry, we might overwrite a normally scheduled downlink, that might
			// then not be relayed by the concentrator...
			ctx.Error("Concentrator is currently emitting")
			return errors.New("Concentrator is already emitting")
		} else if txStatus == C.TX_SCHEDULED {
			ctx.Warn("A downlink was already scheduled, overwriting it")
		}
		break
	}

	concentratorMutex.Lock()
	result := C.lgw_send(txPacket)
	concentratorMutex.Unlock()

	if result == C.LGW_HAL_ERROR {
		ctx.Warn("Downlink transmission to the concentrator failed")
		return errors.New("Downlink transmission to the concentrator failed")
	}

	return nil
}
