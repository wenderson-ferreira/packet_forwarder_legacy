// Copyright © 2017 The Things Network. Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package pktfwd

import (
	"net"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/packet_forwarder/util"
	"github.com/TheThingsNetwork/packet_forwarder/wrapper"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

type StatusManager interface {
	BootTimeSetter
	HandledRXBatch(received, valid int)
	ReceivedTX()
	SentTX()
	GenerateStatus(rtt time.Duration) (*gateway.Status, error)
}

func NewStatusManager(ctx log.Interface, frequencyPlan string, gatewayDescription string, isGPSChip bool, antennaLocation *account.AntennaLocation) StatusManager {
	if antennaLocation == nil {
		ctx.Warn("Antenna location unavailable from the account server")
	}
	return &statusManager{
		antennaLocation:    antennaLocation,
		ctx:                ctx,
		isGPSChip:          isGPSChip,
		rxIn:               0,
		rxOk:               0,
		txIn:               0,
		txOk:               0,
		frequencyPlan:      frequencyPlan,
		gatewayDescription: gatewayDescription,
	}
}

type statusManager struct {
	antennaLocation    *account.AntennaLocation
	ctx                log.Interface
	isGPSChip          bool
	rxIn               uint32
	rxOk               uint32
	txIn               uint32
	txOk               uint32
	frequencyPlan      string
	gatewayDescription string
	bootTime           *time.Time
}

func (s *statusManager) SetBootTime(t time.Time) {
	s.bootTime = &t
}

func (s *statusManager) ReceivedTX() {
	atomic.AddUint32(&s.txIn, 1)
}

func (s *statusManager) SentTX() {
	atomic.AddUint32(&s.txOk, 1)
}

func (s *statusManager) HandledRXBatch(received, valid int) {
	atomic.AddUint32(&s.rxIn, uint32(received))
	atomic.AddUint32(&s.rxOk, uint32(valid))
}

func getOSInfo() *gateway.Status_OSMetrics {
	osInfo := &gateway.Status_OSMetrics{}
	/* Temperature not yet implemented due to disparities between
	platforms (no standard way of getting temperature from a platform
	to another: see https://github.com/shirou/gopsutil/issues/329) */

	stats, err := cpu.Times(false)
	if err == nil && len(stats) > 0 {
		cpuStat := stats[0]
		cpuUsageTime := cpuStat.Total() - cpuStat.Idle
		osInfo.CpuPercentage = float32(cpuUsageTime / cpuStat.Total() * 100)
	} // CPU stats not available on every platform

	loadInfo, err := load.Avg()
	if err == nil {
		osInfo.Load_1 = float32(loadInfo.Load1)
		osInfo.Load_5 = float32(loadInfo.Load5)
		osInfo.Load_15 = float32(loadInfo.Load15)
	}

	virtualMemory, err := mem.VirtualMemory()
	if err == nil {
		osInfo.MemoryPercentage = float32(virtualMemory.UsedPercent)
	}

	return osInfo
}

func (s *statusManager) GenerateStatus(rtt time.Duration) (*gateway.Status, error) {
	var concentratorBootTime time.Duration
	if s.bootTime == nil {
		concentratorBootTime = 0
	} else {
		concentratorBootTime = time.Now().Sub(*s.bootTime)
	}

	osInfo := getOSInfo()
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, errors.Wrap(err, "Net interfaces obtention error")
	}

	ips := make([]string, 0)
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	status := &gateway.Status{
		Timestamp:      uint32(util.TXTimestampFromDuration(concentratorBootTime)),
		Time:           time.Now().UnixNano(),
		GatewayTrusted: true,
		Region:         s.frequencyPlan,
		Ip:             ips,
		Platform:       runtime.GOOS,
		// Contact-email: TODO once it has been implemented on the account server
		ContactEmail: "",
		Description:  s.gatewayDescription,
		Rtt:          uint32(rtt.Nanoseconds() / 1000000),
		RxIn:         atomic.LoadUint32(&s.rxIn),
		RxOk:         atomic.LoadUint32(&s.rxOk),
		TxIn:         atomic.LoadUint32(&s.txIn),
		TxOk:         atomic.LoadUint32(&s.txOk),
		Os:           osInfo,
	}

	coordinates := new(gateway.GPSMetadata)
	if s.antennaLocation != nil { // Antenna location accessible from the account server
		if s.antennaLocation.Latitude != nil {
			coordinates.Latitude = float32(*s.antennaLocation.Latitude)
		}
		if s.antennaLocation.Longitude != nil {
			coordinates.Longitude = float32(*s.antennaLocation.Longitude)
		}
		if s.antennaLocation.Altitude != nil {
			coordinates.Altitude = int32(*s.antennaLocation.Altitude)
		}
	}

	if s.isGPSChip { // GPS chip available
		gpsChipCoordinates, err := wrapper.GetGPSCoordinates()
		if err != nil {
			s.ctx.WithError(err).Warn("Unable to retrieve GPS coordinates from the GPS hardware")
		} else {
			coordinates = &gateway.GPSMetadata{
				Latitude:  float32(gpsChipCoordinates.Latitude),
				Longitude: float32(gpsChipCoordinates.Longitude),
				Altitude:  int32(gpsChipCoordinates.Altitude),
			}
		}
	}

	status.Gps = coordinates

	return status, nil
}
