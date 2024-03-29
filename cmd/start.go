// Copyright © 2017 The Things Network. Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"os"
	"runtime/trace"
	"strconv"
	"time"

	"github.com/TheThingsNetwork/packet_forwarder/pktfwd"
	"github.com/TheThingsNetwork/packet_forwarder/util"
	"github.com/TheThingsNetwork/packet_forwarder/wrapper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// standardDownlinkSendMargin is the time we send a TX packet to the concentrator before its sending time.
const standardDownlinkSendMargin = 100

// downlinksMargin is specified at build. If it contains a numeric value, it is used as the number of
// milliseconds of time margin. If no numeric value can be parsed, we use standardTimeMargin.
var downlinksSendMargin = ""

func getDefaultDownlinkSendMargin() int64 {
	margin, err := strconv.Atoi(downlinksSendMargin)
	if err != nil || margin == 0 {
		return standardDownlinkSendMargin
	}

	return int64(margin)
}

var config = viper.GetViper()

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Packet Forwarding",
	Long:  `packet-forwarder start connects to the LoRa concentrator, and starts redirecting the packets.`,

	Run: func(cmd *cobra.Command, args []string) {
		ctx := util.GetLogger()
		ctx.WithField("HALVersionInfo", wrapper.LoRaGatewayVersionInfo()).Info("Packet Forwarder for LoRa Gateway")

		if traceFilename := config.GetString("run-trace"); traceFilename != "" {
			f, err := os.Create(traceFilename)
			if err != nil {
				ctx.WithField("File", traceFilename).Fatal("Couldn't create trace file")
			}
			trace.Start(f)
			defer trace.Stop()
			ctx.WithField("File", traceFilename).Info("Trace writing active for this run")
		}

		if pin := config.GetInt("reset-pin"); pin != 0 {
			ctx.WithField("ResetPin", pin).Info("Reset pin specified, resetting concentrator...")
			if err := pktfwd.ResetPin(pin); err != nil {
				ctx.WithError(err).Fatal("Couldn't reset pin")
			}
		}

		ignoreCRC := config.GetBool("ignore-crc")
		if ignoreCRC {
			ctx.Warn("CRC check disabled, packets with invalid CRC will be sent upstream")
		}

		ttnConfig := &pktfwd.TTNConfig{
			ID:                  config.GetString("id"),
			Key:                 config.GetString("key"),
			AuthServer:          config.GetString("auth-server"),
			DiscoveryServer:     config.GetString("discovery-server"),
			Router:              config.GetString("router"),
			Version:             config.GetString("version"),
			DownlinksSendMargin: time.Duration(config.GetInt64("downlink-send-margin")) * time.Millisecond,
			IgnoreCRC:           ignoreCRC,
		}

		conf, err := pktfwd.FetchConfig(ctx, ttnConfig)
		if err != nil {
			ctx.WithError(err).Fatal("Couldn't read configuration")
			return
		}

		if err = pktfwd.Run(ctx, *conf, *ttnConfig, config.GetString("gps-path")); err != nil {
			ctx.WithError(err).Error("The program ended following a failure")
		}
	},
}

func init() {
	startCmd.PersistentFlags().String("auth-server", "https://account.thethingsnetwork.org", "The account server the packet forwarder gets the gateway configuration from")
	startCmd.PersistentFlags().String("discovery-server", "discover.thethingsnetwork.org:1900", "The discovery server the packet forwarder uses to route the packets")
	startCmd.PersistentFlags().String("id", "", "The gateway ID to get its configuration from the account server")
	startCmd.PersistentFlags().String("key", "", "The gateway key to authenticate itself with the back-end")
	startCmd.PersistentFlags().String("router", "", "The router to communicate with (example: ttn-router-eu)")
	startCmd.PersistentFlags().String("gps-path", "", "The file system path to the GPS interface, if a GPS is available (example: /dev/nmea)")
	startCmd.PersistentFlags().Int64("downlink-send-margin", getDefaultDownlinkSendMargin(), "The margin, in milliseconds, between a downlink is sent to a concentrator and it is being sent by the concentrator")
	startCmd.PersistentFlags().String("run-trace", "", "File to which write the runtime trace of the packet forwarder. Can later be read with `go tool trace <trace_file>`.")
	startCmd.PersistentFlags().Int("reset-pin", 0, "GPIO pin associated to the reset pin of the board")
	startCmd.PersistentFlags().BoolP("verbose", "v", false, "Show debug logs")
	startCmd.PersistentFlags().Bool("ignore-crc", false, "Send packets upstream even if CRC validation is incorrect")

	viper.BindPFlags(startCmd.PersistentFlags())

	RootCmd.AddCommand(startCmd)
}
