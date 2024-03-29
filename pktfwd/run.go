// Copyright © 2017 The Things Network. Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package pktfwd

import (
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/packet_forwarder/util"
	"github.com/pkg/errors"
)

// Init initiates the configuration, the network connection, and handles the manager
func Run(ctx log.Interface, conf util.Config, ttnConfig TTNConfig, gpsPath string) error {
	networkCli, err := CreateNetworkClient(ctx, ttnConfig)
	if err != nil {
		return errors.Wrap(err, "Network configuration failure")
	}

	// applying configuration to the board
	if err := configureBoard(ctx, conf, gpsPath); err != nil {
		return errors.Wrap(err, "Board configuration failure")
	}

	// Creating manager
	var mgr = NewManager(ctx, conf, networkCli, gpsPath, ttnConfig)
	return mgr.run()
}
