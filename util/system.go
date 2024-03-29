// Copyright © 2017 The Things Network. Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"os"
	"path"
	"time"

	"github.com/spf13/viper"
)

// TXTimestamp allows to wrap a router.DownlinkMessage.GatewayConfiguration.Timestamp
type TXTimestamp uint32

func (ts TXTimestamp) GetAsDuration() time.Duration {
	return time.Duration(ts) * time.Microsecond
}

func TXTimestampFromDuration(d time.Duration) TXTimestamp {
	return TXTimestamp(d.Nanoseconds() / 1000.0)
}

func GetConfigFile() string {
	flag := viper.GetString("config")

	home := os.Getenv("HOME")
	homeyml := ""
	homeyaml := ""

	if home != "" {
		homeyml = path.Join(home, ".pktfwd.yml")
		homeyaml = path.Join(home, ".pktfwd.yaml")
	}

	try_files := []string{
		flag,
		homeyml,
		homeyaml,
	}

	// find a file that exists, and use that
	for _, file := range try_files {
		if file != "" {
			if _, err := os.Stat(file); err == nil {
				return file
			}
		}
	}

	// no file found, set up correct fallback
	return homeyml
}
