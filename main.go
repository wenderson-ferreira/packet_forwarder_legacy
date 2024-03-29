// Copyright © 2017 The Things Network. Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"github.com/TheThingsNetwork/packet_forwarder/cmd"
	"github.com/spf13/viper"
)

var (
	version   = "2.x.x"
	gitCommit = "unknown"
	buildDate = "unknown"
)

func main() {
	viper.Set("version", version)
	viper.Set("gitCommit", gitCommit)
	viper.Set("buildDate", buildDate)
	cmd.Execute()
}
