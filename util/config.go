// Copyright © 2017 The Things Network. Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"io/ioutil"
	"net/http"

	"encoding/json"

	"github.com/TheThingsNetwork/go-utils/log"
)

type ChannelConf struct {
	Enabled      bool    `json:"enable"`
	Description  *string `json:"desc,omitempty"`
	Radio        uint8   `json:"radio"`
	IfValue      int32   `json:"if"`
	Bandwidth    *uint32 `json:"bandwidth,omitempty"`
	Datarate     *uint32 `json:"datarate,omitempty"`
	SpreadFactor *uint8  `json:"spread_factor,omitempty"`
}

type ChannelFreqConf struct {
	Freq     int `json:"freq_hz"` // Frequency in hertz
	ScanTime int `json:"scan_time_us"`
}

// LBTConf wraps lbt configuration for SX1301
type LbtConf struct {
	Enabled        bool              `json:"enable"`
	RssiTarget     int               `json:"rssi_target"`
	RssiOffset     int               `json:"sx127x_rssi_offset"`
	ChannelsConfig []ChannelFreqConf `json:"chan_cfg"`
}

type GainTableConf struct {
	PaGain      uint8   `json:"pa_gain"`
	MixGain     uint8   `json:"mix_gain"`
	RfPower     int8    `json:"rf_power"`
	DigGain     uint8   `json:"dig_gain"`
	Description *string `json:"desc,omitempty"`
	DacGain     *uint8  `json:"dac_gain,omitempty"`
}

type RadioConf struct {
	Enabled     bool    `json:"enable"`
	RadioType   string  `json:"type"`
	Freq        int     `json:"freq"`
	RssiOffset  float32 `json:"rssi_offset"`
	TxEnabled   bool    `json:"tx_enable"`
	TxNotchFreq *int    `json:"tx_notch_freq,omitempty"`
	TxMinFreq   *int    `json:"tx_freq_min,omitempty"`
	TxMaxFreq   *int    `json:"tx_freq_max,omitempty"`
}

type SX1301Conf struct {
	LorawanPublic          bool           `json:"lorawan_public"`
	Clksrc                 int            `json:"clksrc"`
	ClksrcDescription      *string        `json:"clksrc_desc,omitempty"`
	AntennaGain            *int           `json:"antenna_gain,omitempty"`
	AntennaGainDescription *string        `json:"antenna_gain_desc,omitempty"`
	LbtConfig              *LbtConf       `json:"lbt_cfg,omitempty"`
	Radio0                 *RadioConf     `json:"radio_0,omitempty"`
	Radio1                 *RadioConf     `json:"radio_1,omitempty"`
	MultiSFChan0           *ChannelConf   `json:"chan_multiSF_0,omitempty"`
	MultiSFChan1           *ChannelConf   `json:"chan_multiSF_1,omitempty"`
	MultiSFChan2           *ChannelConf   `json:"chan_multiSF_2,omitempty"`
	MultiSFChan3           *ChannelConf   `json:"chan_multiSF_3,omitempty"`
	MultiSFChan4           *ChannelConf   `json:"chan_multiSF_4,omitempty"`
	MultiSFChan5           *ChannelConf   `json:"chan_multiSF_5,omitempty"`
	MultiSFChan6           *ChannelConf   `json:"chan_multiSF_6,omitempty"`
	MultiSFChan7           *ChannelConf   `json:"chan_multiSF_7,omitempty"`
	MultiSFChan8           *ChannelConf   `json:"chan_multiSF_8,omitempty"`
	MultiSFChan9           *ChannelConf   `json:"chan_multiSF_9,omitempty"`
	MultiSFChan10          *ChannelConf   `json:"chan_multiSF_10,omitempty"`
	MultiSFChan11          *ChannelConf   `json:"chan_multiSF_11,omitempty"`
	MultiSFChan12          *ChannelConf   `json:"chan_multiSF_12,omitempty"`
	MultiSFChan13          *ChannelConf   `json:"chan_multiSF_13,omitempty"`
	MultiSFChan14          *ChannelConf   `json:"chan_multiSF_14,omitempty"`
	MultiSFChan15          *ChannelConf   `json:"chan_multiSF_15,omitempty"`
	MultiSFChan16          *ChannelConf   `json:"chan_multiSF_16,omitempty"`
	MultiSFChan17          *ChannelConf   `json:"chan_multiSF_17,omitempty"`
	MultiSFChan18          *ChannelConf   `json:"chan_multiSF_18,omitempty"`
	MultiSFChan19          *ChannelConf   `json:"chan_multiSF_19,omitempty"`
	MultiSFChan20          *ChannelConf   `json:"chan_multiSF_20,omitempty"`
	MultiSFChan21          *ChannelConf   `json:"chan_multiSF_21,omitempty"`
	MultiSFChan22          *ChannelConf   `json:"chan_multiSF_22,omitempty"`
	MultiSFChan23          *ChannelConf   `json:"chan_multiSF_23,omitempty"`
	MultiSFChan24          *ChannelConf   `json:"chan_multiSF_24,omitempty"`
	MultiSFChan25          *ChannelConf   `json:"chan_multiSF_25,omitempty"`
	MultiSFChan26          *ChannelConf   `json:"chan_multiSF_26,omitempty"`
	MultiSFChan27          *ChannelConf   `json:"chan_multiSF_27,omitempty"`
	MultiSFChan28          *ChannelConf   `json:"chan_multiSF_28,omitempty"`
	MultiSFChan29          *ChannelConf   `json:"chan_multiSF_29,omitempty"`
	MultiSFChan30          *ChannelConf   `json:"chan_multiSF_30,omitempty"`
	MultiSFChan31          *ChannelConf   `json:"chan_multiSF_31,omitempty"`
	MultiSFChan32          *ChannelConf   `json:"chan_multiSF_32,omitempty"`
	MultiSFChan33          *ChannelConf   `json:"chan_multiSF_33,omitempty"`
	MultiSFChan34          *ChannelConf   `json:"chan_multiSF_34,omitempty"`
	MultiSFChan35          *ChannelConf   `json:"chan_multiSF_35,omitempty"`
	MultiSFChan36          *ChannelConf   `json:"chan_multiSF_36,omitempty"`
	MultiSFChan37          *ChannelConf   `json:"chan_multiSF_37,omitempty"`
	MultiSFChan38          *ChannelConf   `json:"chan_multiSF_38,omitempty"`
	MultiSFChan39          *ChannelConf   `json:"chan_multiSF_39,omitempty"`
	MultiSFChan40          *ChannelConf   `json:"chan_multiSF_40,omitempty"`
	MultiSFChan41          *ChannelConf   `json:"chan_multiSF_41,omitempty"`
	MultiSFChan42          *ChannelConf   `json:"chan_multiSF_42,omitempty"`
	MultiSFChan43          *ChannelConf   `json:"chan_multiSF_43,omitempty"`
	MultiSFChan44          *ChannelConf   `json:"chan_multiSF_44,omitempty"`
	MultiSFChan45          *ChannelConf   `json:"chan_multiSF_45,omitempty"`
	MultiSFChan46          *ChannelConf   `json:"chan_multiSF_46,omitempty"`
	MultiSFChan47          *ChannelConf   `json:"chan_multiSF_47,omitempty"`
	MultiSFChan48          *ChannelConf   `json:"chan_multiSF_48,omitempty"`
	MultiSFChan49          *ChannelConf   `json:"chan_multiSF_49,omitempty"`
	MultiSFChan50          *ChannelConf   `json:"chan_multiSF_50,omitempty"`
	MultiSFChan51          *ChannelConf   `json:"chan_multiSF_51,omitempty"`
	MultiSFChan52          *ChannelConf   `json:"chan_multiSF_52,omitempty"`
	MultiSFChan53          *ChannelConf   `json:"chan_multiSF_53,omitempty"`
	MultiSFChan54          *ChannelConf   `json:"chan_multiSF_54,omitempty"`
	MultiSFChan55          *ChannelConf   `json:"chan_multiSF_55,omitempty"`
	MultiSFChan56          *ChannelConf   `json:"chan_multiSF_56,omitempty"`
	MultiSFChan57          *ChannelConf   `json:"chan_multiSF_57,omitempty"`
	MultiSFChan58          *ChannelConf   `json:"chan_multiSF_58,omitempty"`
	MultiSFChan59          *ChannelConf   `json:"chan_multiSF_59,omitempty"`
	MultiSFChan60          *ChannelConf   `json:"chan_multiSF_60,omitempty"`
	MultiSFChan61          *ChannelConf   `json:"chan_multiSF_61,omitempty"`
	MultiSFChan62          *ChannelConf   `json:"chan_multiSF_62,omitempty"`
	MultiSFChan63          *ChannelConf   `json:"chan_multiSF_63,omitempty"`
	LoraSTDChannel         *ChannelConf   `json:"chan_Lora_std,omitempty"`
	FSKChannel             *ChannelConf   `json:"chan_FSK,omitempty"`
	TxLut0                 *GainTableConf `json:"tx_lut_0,omitempty"`
	TxLut1                 *GainTableConf `json:"tx_lut_1,omitempty"`
	TxLut2                 *GainTableConf `json:"tx_lut_2,omitempty"`
	TxLut3                 *GainTableConf `json:"tx_lut_3,omitempty"`
	TxLut4                 *GainTableConf `json:"tx_lut_4,omitempty"`
	TxLut5                 *GainTableConf `json:"tx_lut_5,omitempty"`
	TxLut6                 *GainTableConf `json:"tx_lut_6,omitempty"`
	TxLut7                 *GainTableConf `json:"tx_lut_7,omitempty"`
	TxLut8                 *GainTableConf `json:"tx_lut_8,omitempty"`
	TxLut9                 *GainTableConf `json:"tx_lut_9,omitempty"`
	TxLut10                *GainTableConf `json:"tx_lut_10,omitempty"`
	TxLut11                *GainTableConf `json:"tx_lut_11,omitempty"`
	TxLut12                *GainTableConf `json:"tx_lut_12,omitempty"`
	TxLut13                *GainTableConf `json:"tx_lut_13,omitempty"`
	TxLut14                *GainTableConf `json:"tx_lut_14,omitempty"`
	TxLut15                *GainTableConf `json:"tx_lut_15,omitempty"`
}

func (s SX1301Conf) GetRadios() []RadioConf {
	radios := make([]RadioConf, 0)
	for _, i := range []*RadioConf{s.Radio0, s.Radio1} {
		if i == nil {
			return radios
		}
		radios = append(radios, *i)
	}
	return radios
}

func (s SX1301Conf) GetTXLuts() []GainTableConf {
	gainTables := make([]GainTableConf, 0)
	for _, i := range []*GainTableConf{
		s.TxLut0, s.TxLut1, s.TxLut2, s.TxLut3, s.TxLut4, s.TxLut5, s.TxLut6, s.TxLut7, s.TxLut8, s.TxLut9,
		s.TxLut10, s.TxLut11, s.TxLut12, s.TxLut13, s.TxLut14, s.TxLut15,
	} {
		if i == nil {
			return gainTables
		}
		gainTables = append(gainTables, *i)
	}
	return gainTables
}

func (s SX1301Conf) GetMultiSFChannels() []ChannelConf {
	channels := make([]ChannelConf, 0)
	for _, i := range []*ChannelConf{
		s.MultiSFChan0, s.MultiSFChan1, s.MultiSFChan2, s.MultiSFChan3, s.MultiSFChan4, s.MultiSFChan5, s.MultiSFChan6, s.MultiSFChan7, s.MultiSFChan8, s.MultiSFChan9,
		s.MultiSFChan10, s.MultiSFChan11, s.MultiSFChan12, s.MultiSFChan13, s.MultiSFChan14, s.MultiSFChan15, s.MultiSFChan16, s.MultiSFChan17, s.MultiSFChan18, s.MultiSFChan19,
		s.MultiSFChan20, s.MultiSFChan21, s.MultiSFChan22, s.MultiSFChan23, s.MultiSFChan24, s.MultiSFChan25, s.MultiSFChan26, s.MultiSFChan27, s.MultiSFChan28, s.MultiSFChan29,
		s.MultiSFChan30, s.MultiSFChan31, s.MultiSFChan32, s.MultiSFChan33, s.MultiSFChan34, s.MultiSFChan35, s.MultiSFChan36, s.MultiSFChan37, s.MultiSFChan38, s.MultiSFChan39,
		s.MultiSFChan40, s.MultiSFChan41, s.MultiSFChan42, s.MultiSFChan43, s.MultiSFChan44, s.MultiSFChan45, s.MultiSFChan46, s.MultiSFChan47, s.MultiSFChan48, s.MultiSFChan49,
		s.MultiSFChan50, s.MultiSFChan51, s.MultiSFChan52, s.MultiSFChan53, s.MultiSFChan54, s.MultiSFChan55, s.MultiSFChan56, s.MultiSFChan57, s.MultiSFChan58, s.MultiSFChan59,
		s.MultiSFChan60, s.MultiSFChan61, s.MultiSFChan62, s.MultiSFChan63,
	} {
		if i == nil {
			return channels
		}
		channels = append(channels, *i)
	}
	return channels
}

type Config struct {
	Concentrator SX1301Conf `json:"SX1301_conf"`
}

func jsonParseConfig(frequencyPlan []byte) (Config, error) {
	conf := Config{}
	if err := json.Unmarshal(frequencyPlan, &conf); err != nil {
		return conf, err
	}
	return conf, nil
}

func FetchConfigFromURL(ctx log.Interface, url string) (Config, error) {
	c := Config{}

	resp, err := http.Get(url)
	if err != nil {
		ctx.Error("Couldn't get the frequency plans")
		return c, err
	}
	frequency, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ctx.Error("Failure to read the server response")
		return c, err
	}
	resp.Body.Close()

	return jsonParseConfig(frequency)
}
