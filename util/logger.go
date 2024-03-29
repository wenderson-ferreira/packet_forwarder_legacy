// Copyright © 2017 The Things Network. Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"os"

	cliHandler "github.com/TheThingsNetwork/go-utils/handlers/cli"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/log/apex"
	"github.com/apex/log"
	levelHandler "github.com/apex/log/handlers/level"
	"github.com/spf13/viper"
)

func GetLogger() ttnlog.Interface {
	logLevel := log.InfoLevel
	if viper.GetBool("verbose") {
		logLevel = log.DebugLevel
	}
	ctx := apex.Wrap(&log.Logger{
		Handler: levelHandler.New(cliHandler.New(os.Stdout), logLevel),
	})
	return ctx
}
