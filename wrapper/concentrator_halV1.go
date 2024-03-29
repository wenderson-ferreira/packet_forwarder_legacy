// +build halv1

package wrapper

// #cgo CFLAGS: -I${SRCDIR}/../lora_gateway/libloragw/inc
// #cgo LDFLAGS: -lm ${SRCDIR}/../lora_gateway/libloragw/libloragw.a
// #include "config.h"
// #include "loragw_hal.h"
// #include "loragw_gps.h"
// void setType(struct lgw_conf_rxrf_s *rxrfConf, enum lgw_radio_type_e val) {
// 	rxrfConf->type = val;
// }
import "C"
import (
	"errors"
	"fmt"
	"sync"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/packet_forwarder/util"
)

var concentratorMutex = &sync.Mutex{}

var loraChannelBandwidths = map[uint32]C.uint8_t{
	7800:   C.BW_7K8HZ,
	15600:  C.BW_15K6HZ,
	31200:  C.BW_31K2HZ,
	62500:  C.BW_62K5HZ,
	125000: C.BW_125KHZ,
	250000: C.BW_250KHZ,
	500000: C.BW_500KHZ,
}

var loraChannelSpreadingFactors = map[uint32]C.uint32_t{
	7:  C.DR_LORA_SF7,
	8:  C.DR_LORA_SF8,
	9:  C.DR_LORA_SF9,
	10: C.DR_LORA_SF10,
	11: C.DR_LORA_SF11,
	12: C.DR_LORA_SF12,
}

// LoRaGatewayVersionInfo returns a string with information on the HAL
func LoRaGatewayVersionInfo() string {
	var versionInfo = C.GoString(C.lgw_version_info())
	return versionInfo
}

// StartLoRaGateway wraps the HAL function to start the concentrator once configured
func StartLoRaGateway() error {
	state := C.lgw_start()

	if state != C.LGW_HAL_SUCCESS {
		return errors.New("Failed to start concentrator")
	}
	return nil
}

// StopLoRaGateway wraps the HAL function to stop the concentrator once started
func StopLoRaGateway() error {
	state := C.lgw_stop()

	if state != C.LGW_HAL_SUCCESS {
		return errors.New("Failed to stop concentrator gracefully")
	}
	return nil
}

// SetBoardConf wraps the HAL function to configure the concentrator's board
func SetBoardConf(ctx log.Interface, conf util.Config) error {
	var boardConf = C.struct_lgw_conf_board_s{
		clksrc:         C.uint8_t(conf.Concentrator.Clksrc),
		lorawan_public: C.bool(conf.Concentrator.LorawanPublic),
	}

	if C.lgw_board_setconf(boardConf) != C.LGW_HAL_SUCCESS {
		return errors.New("Failed board configuration")
	}
	ctx.WithFields(log.Fields{
		"ClockSource":   conf.Concentrator.Clksrc,
		"LorawanPublic": conf.Concentrator.LorawanPublic,
	}).Info("SX1301 board configured")
	return nil
}

/* prepareTXLut takes the pointer to an empty C.struct_lgw_tx_gain_s, its configuration wrapped in Go, and transposes
the configuration in the C.struct_lgw_tx_gain_s. It also increments the size of the TX Gain Lut table. */
func prepareTXLut(txLut *C.struct_lgw_tx_gain_s, txConf util.GainTableConf) {
	if txConf.DacGain != nil {
		txLut.dac_gain = C.uint8_t(*txConf.DacGain)
	} else {
		txLut.dac_gain = 3
	}
	txLut.dig_gain = C.uint8_t(txConf.DigGain)
	txLut.mix_gain = C.uint8_t(txConf.MixGain)
	txLut.rf_power = C.int8_t(txConf.RfPower)
	txLut.pa_gain = C.uint8_t(txConf.PaGain)
}

// SetTXGainConf prepares, and then sends the configuration of the TX Gain LUT to the concentrator
func SetTXGainConf(ctx log.Interface, conc util.SX1301Conf) error {
	var gainLut = C.struct_lgw_tx_gain_lut_s{
		size: 0,
		lut:  [C.TX_GAIN_LUT_SIZE_MAX]C.struct_lgw_tx_gain_s{},
	}
	txLuts := conc.GetTXLuts()
	for i, txLut := range txLuts {
		prepareTXLut(&gainLut.lut[i], txLut)
	}
	gainLut.size = C.uint8_t(len(txLuts))

	if C.lgw_txgain_setconf(&gainLut) != C.LGW_HAL_SUCCESS {
		return errors.New("Failed to configure concentrator TX Gain LUT")
	}
	ctx.WithField("Indexes", gainLut.size).Info("Configured TX Lut")
	return nil
}

// initRadio initiates a radio configuration in the C.struct_lgw_conf_rxrf_s format, given
// the configuration for that radio.
func initRadio(radio util.RadioConf) (C.struct_lgw_conf_rxrf_s, error) {
	var cRadio = C.struct_lgw_conf_rxrf_s{
		enable:      C.bool(radio.Enabled),
		freq_hz:     C.uint32_t(radio.Freq),
		rssi_offset: C.float(radio.RssiOffset),
		tx_enable:   C.bool(radio.TxEnabled),
	}

	// Checking the radio is of a pre-defined type
	switch radio.RadioType {
	case "SX1257":
		C.setType(&cRadio, C.LGW_RADIO_TYPE_SX1257)
	case "SX1255":
		C.setType(&cRadio, C.LGW_RADIO_TYPE_SX1255)
	default:
		return cRadio, errors.New("Invalid radio type (should be SX1255 or SX1257)")
	}
	return cRadio, nil
}

// enableRadio is enabling the radio
func enableRadio(ctx log.Interface, radio util.RadioConf, nb uint8) error {
	// Checking if radio is enabled and thus needs to be activated
	if !radio.Enabled {
		ctx.WithField("Radio", nb).Info("Radio disabled")
		return nil
	}

	cRadio, err := initRadio(radio)
	if err != nil {
		return err
	}

	if C.lgw_rxrf_setconf(C.uint8_t(nb), cRadio) != C.LGW_HAL_SUCCESS {
		ctx.WithField("Radio", nb).Warn("Invalid configuration")
		return errors.New("Radio configuration failed")
	}

	ctx.WithFields(log.Fields{
		"Radio":      nb,
		"Type":       radio.RadioType,
		"EnabledTX":  radio.TxEnabled,
		"Frequency":  radio.Freq,
		"RSSIOffset": radio.RssiOffset,
	}).Info("Radio configured")
	return nil
}

// SetRFChannels send the configuration of the radios to the concentrator
func SetRFChannels(ctx log.Interface, conf util.Config) error {
	for i, radio := range conf.Concentrator.GetRadios() {
		err := enableRadio(ctx, radio, uint8(i))
		if err != nil {
			return err
		}
	}

	return nil
}

func enableSFChannel(ctx log.Interface, channelConf util.ChannelConf, nb uint8) error {
	if !channelConf.Enabled {
		ctx.WithField("Channel", nb).Info("Lora multi-SF channel disabled")
		return nil
	}

	var cChannel = C.struct_lgw_conf_rxif_s{
		enable:   C.bool(channelConf.Enabled),
		rf_chain: C.uint8_t(channelConf.Radio),
		freq_hz:  C.int32_t(channelConf.IfValue),
	}

	channelLog := ctx.WithField("Lora multi-SF channel", nb)
	if C.lgw_rxif_setconf(C.uint8_t(nb), cChannel) != C.LGW_HAL_SUCCESS {
		return errors.New(fmt.Sprintf("Missing configuration for SF channel %d", nb))
	}
	channelLog.WithFields(log.Fields{
		"RFChain": channelConf.Radio,
		"Freq":    channelConf.IfValue,
	}).Info("LoRa multi-SF channel configured")
	return nil
}

// SetSFChannels enables the different SF channels
func SetSFChannels(ctx log.Interface, conf util.Config) error {
	for i, sfChannel := range conf.Concentrator.GetMultiSFChannels() {
		err := enableSFChannel(ctx, sfChannel, uint8(i))
		if err != nil {
			return err
		}
	}
	return nil
}

// initLoRaStdChannel initiates a C.struct_lgw_conf_rxif_s from a LoRaChannelConf
func initLoRaStdChannel(stdChan util.ChannelConf) C.struct_lgw_conf_rxif_s {
	var cChannel = C.struct_lgw_conf_rxif_s{
		enable:   C.bool(stdChan.Enabled),
		rf_chain: C.uint8_t(stdChan.Radio),
		freq_hz:  C.int32_t(stdChan.IfValue),
	}

	switch *stdChan.Bandwidth {
	case 125000, 250000, 500000:
		cChannel.bandwidth = loraChannelBandwidths[*stdChan.Bandwidth]
	default:
		cChannel.bandwidth = C.BW_UNDEFINED
	}

	if stdChan.Datarate != nil && *stdChan.Datarate >= 7 && *stdChan.Datarate <= 12 {
		cChannel.datarate = loraChannelSpreadingFactors[*stdChan.Datarate]
	} else {
		cChannel.datarate = C.DR_UNDEFINED
	}

	return cChannel
}

// SetStandardChannel enables the LoRa standard channel from the configuration
func SetStandardChannel(ctx log.Interface, stdChan util.ChannelConf) error {
	if !stdChan.Enabled {
		ctx.Info("LoRa standard channel disabled")
		return nil
	}

	var cChannel = initLoRaStdChannel(stdChan)

	if C.lgw_rxif_setconf(8, cChannel) != C.LGW_HAL_SUCCESS {
		return errors.New("Configuration for LoRa standard channel failed")
	}
	return nil
}

// SetFSKChannel sets the FSK Channel configuration on the concentrator
func SetFSKChannel(ctx log.Interface, fskChan util.ChannelConf) error {
	if !fskChan.Enabled {
		ctx.Info("FSK channel disabled")
		return nil
	}

	var cFSKChan = C.struct_lgw_conf_rxif_s{
		enable:    C.bool(fskChan.Enabled),
		rf_chain:  C.uint8_t(fskChan.Radio),
		freq_hz:   C.int32_t(fskChan.IfValue),
		bandwidth: C.BW_UNDEFINED,
	}
	if fskChan.Datarate != nil {
		cFSKChan.datarate = C.uint32_t(*fskChan.Datarate)
	}
	if fskChan.Bandwidth == nil {
		return errors.New("No bandwidth information in the configuration for the FSK channel - cannot retransmit the FSK packet")
	}

	val := *fskChan.Bandwidth
	switch {
	case val > 0 && val <= 7800:
		cFSKChan.bandwidth = loraChannelBandwidths[7800]
	case val > 7800 && val <= 15600:
		cFSKChan.bandwidth = loraChannelBandwidths[15600]
	case val > 15600 && val <= 31200:
		cFSKChan.bandwidth = loraChannelBandwidths[31200]
	case val > 31200 && val <= 62500:
		cFSKChan.bandwidth = loraChannelBandwidths[62500]
	case val > 62500 && val <= 125000:
		cFSKChan.bandwidth = loraChannelBandwidths[125000]
	case val > 125000 && val <= 250000:
		cFSKChan.bandwidth = loraChannelBandwidths[250000]
	case val > 250000 && val <= 500000:
		cFSKChan.bandwidth = loraChannelBandwidths[500000]
	}

	if C.lgw_rxif_setconf(9, cFSKChan) != C.LGW_HAL_SUCCESS {
		return errors.New("Configuration for FSK channel failed")
	}
	return nil
}
