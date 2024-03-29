// Copyright © 2017 The Things Network. Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package pktfwd

import (
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/packet_forwarder/wrapper"
	"github.com/pkg/errors"
)

/*
	GPS workflow with the TTN back-end:
	- if gps available, send that in the status message (no need to send it with every uplink)
	- if nothing available or set, get the coordinates from account server and use that in the status message
*/

// enableGPS checks if there is an available GPS for this build - if yes,
// tries to activate it.
func enableGPS(ctx log.Interface, gpsPath string) (err error) {
	if gpsPath == "" {
		ctx.Warn("No GPS chip configured, ignoring")
		return nil
	}

	ctx.WithField("GPSPath", gpsPath).Info("GPS path found, activating")
	err = wrapper.LoRaGPSEnable(gpsPath)
	if err != nil {
		return errors.Wrap(err, "GPS activation failed")
	}

	return nil
}
