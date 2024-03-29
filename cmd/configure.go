// Copyright © 2017 The Things Network. Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cmd

import (
	"os"

	"github.com/TheThingsNetwork/packet_forwarder/util"
	"github.com/segmentio/go-prompt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var configureCmd = &cobra.Command{
	Use:   "configure [config-path]",
	Short: "Configure Packet Forwarder",
	Long: `packet-forwarder configure creates a YAML configuration file for the packet forwarder.

The first argument is used as the storage location to the configuration file. If nothing is specified, the default configuration file path ($HOME/.pktfwd.yml) is used.`,

	Run: func(cmd *cobra.Command, args []string) {
		ctx := util.GetLogger()

		ctx.Info("If you haven't registered your gateway yet, you can register it either with the console, or with `ttnctl`.")

		var (
			gatewayAuthServer      string
			gatewayDiscoveryServer string
		)

		if !prompt.Confirm("Is this gateway going to be used on the community network?") {
			gatewayDiscoveryServer = prompt.StringRequired("Enter the URL of the discovery server of your private network, in a <ip:port> format:")
			if prompt.Confirm("Are you using a private account server?") {
				gatewayAuthServer = prompt.StringRequired("Enter the URL of the account server (example: \"https://account.thethingsnetwork.org\"")
			}
		}

		gatewayID := prompt.StringRequired("Enter the ID of the gateway")
		gatewayKey := prompt.PasswordMasked("Enter the access key of the gateway")

		type yamlConfig struct {
			ID              string `yaml:"id"`
			Key             string `yaml:"key"`
			AuthServer      string `yaml:"auth-server,omitempty"`
			DiscoveryServer string `yaml:"discovery-server,omitempty"`
		}

		newConfig := &yamlConfig{
			ID:              gatewayID,
			Key:             gatewayKey,
			AuthServer:      gatewayAuthServer,
			DiscoveryServer: gatewayDiscoveryServer,
		}

		output, err := yaml.Marshal(newConfig)
		if err != nil {
			util.GetLogger().WithError(err).Fatal("Failed to generate YAML")
		}

		f, err := os.Create(cfgFile)
		if err != nil {
			util.GetLogger().WithError(err).Fatal("Failed to create file")
		}
		defer f.Close()

		f.Write(output)
		ctx.WithField("ConfigFilePath", cfgFile).Info("New configuration file saved")
	},
}

func init() {
	RootCmd.AddCommand(configureCmd)
}
