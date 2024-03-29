// +build dummy

package wrapper

import (
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/packet_forwarder/util"
	"github.com/TheThingsNetwork/ttn/api/router"
)

func SendDownlink(downlink *router.DownlinkMessage, conf util.Config, ctx log.Interface) error {
	ctx.Info("Dummy HAL - Downlink accepted")
	return nil
}
