// Copyright © 2017 The Things Network. Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package pktfwd

import (
	"github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/packet_forwarder/util"
	"github.com/TheThingsNetwork/packet_forwarder/wrapper"
)

// Multitech concentrators require a clksrc of 0, even if the frequency plan indicates a value of 1.
// This value, modified at build to include the platform type, is currently useful as a flag to
// ignore the frequency plan value of `clksrc`.
var platform = ""

func configureBoard(ctx log.Interface, conf util.Config, gpsPath string) error {
	if platform == "multitech" {
		ctx.Info("Forcing clock source to 0 (Multitech concentrator)")
		conf.Concentrator.Clksrc = 0
	}

	err := wrapper.SetBoardConf(ctx, conf)
	if err != nil {
		return err
	}

	err = configureChannels(ctx, conf)
	if err != nil {
		return err
	}

	err = enableGPS(ctx, gpsPath)
	if err != nil {
		return err
	}

	return nil
}

func configureIndividualChannels(ctx log.Interface, conf util.Config) error {
	// Configuring LoRa standard channel
	if lora := conf.Concentrator.LoraSTDChannel; lora != nil {
		err := wrapper.SetStandardChannel(ctx, *lora)
		if err != nil {
			return err
		}
		ctx.Info("LoRa standard channel configured")
	} else {
		ctx.Warn("No configuration for LoRa standard channel, ignoring")
	}

	// Configuring FSK channel
	if fsk := conf.Concentrator.FSKChannel; fsk != nil {
		err := wrapper.SetFSKChannel(ctx, *fsk)
		if err != nil {
			return err
		}
		ctx.Info("FSK channel configured")
	} else {
		ctx.Warn("No configuration for FSK standard channel, ignoring")
	}

	return nil
}

func configureChannels(ctx log.Interface, conf util.Config) error {
	// Configuring the TX Gain Lut
	err := wrapper.SetTXGainConf(ctx, conf.Concentrator)
	if err != nil {
		return err
	}

	// Configuring the RF and SF channels
	err = wrapper.SetRFChannels(ctx, conf)
	if err != nil {
		return err
	}
	wrapper.SetSFChannels(ctx, conf)

	// Configuring the individual LoRa standard and FSK channels
	err = configureIndividualChannels(ctx, conf)
	if err != nil {
		return err
	}
	return nil
}

// FetchConfig reads the configuration from the distant server
func FetchConfig(ctx log.Interface, ttnConfig *TTNConfig) (*util.Config, error) {
	a := account.New(ttnConfig.AuthServer)

	gw, err := a.FindGateway(ttnConfig.ID)
	ctx = ctx.WithFields(log.Fields{"GatewayID": ttnConfig.ID, "AuthServer": ttnConfig.AuthServer})
	if err != nil {
		ctx.WithError(err).Error("Failed to find gateway specified as gateway ID")
		return nil, err
	}
	ctx.WithField("URL", gw.FrequencyPlanURL).Info("Found gateway parameters, getting frequency plans")
	if gw.Attributes.Description != nil {
		ttnConfig.GatewayDescription = *gw.Attributes.Description
	}

	config, err := util.FetchConfigFromURL(ctx, gw.FrequencyPlanURL)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
