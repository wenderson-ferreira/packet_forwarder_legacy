// +build dummy

package wrapper

import (
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/packet_forwarder/util"
)

func LoRaGatewayVersionInfo() string {
	return "Dummy HAL"
}

func StartLoRaGateway() error {
	return nil
}

func StopLoRaGateway() error {
	return nil
}

func SetBoardConf(ctx log.Interface, conf util.Config) error {
	return nil
}

func SetTXGainConf(ctx log.Interface, conc util.SX1301Conf) error {
	return nil
}

func SetRFChannels(ctx log.Interface, conf util.Config) error {
	return nil
}

func SetSFChannels(ctx log.Interface, conf util.Config) error {
	return nil
}

func SetStandardChannel(ctx log.Interface, stdChan util.ChannelConf) error {
	return nil
}

func SetFSKChannel(ctx log.Interface, fskChan util.ChannelConf) error {
	return nil
}
