// Copyright © 2017 The Things Network. Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package pktfwd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/packet_forwarder/util"
	"github.com/TheThingsNetwork/packet_forwarder/wrapper"
	"github.com/pkg/errors"
)

const (
	initUplinkPollingRate   = 100 * time.Microsecond
	stableUplinkPollingRate = 5 * time.Millisecond
	statusRoutineSleepRate  = 15 * time.Second
	gpsUpdateRate           = 5 * time.Millisecond
)

/* Manager struct manages the routines during runtime, once the gateways and network
configuration have been set up. It startes a routine, that it only stopped when the
users wants to close the program or that an error occurs. */
type Manager struct {
	ctx               log.Interface
	conf              util.Config
	netClient         NetworkClient
	statusMgr         StatusManager
	uplinkPollingRate time.Duration
	// Concentrator boot time
	bootTimeSetters     multipleBootTimeSetter
	foundBootTime       bool
	isGPS               bool
	ignoreCRC           bool
	downlinksSendMargin time.Duration
}

func NewManager(ctx log.Interface, conf util.Config, netClient NetworkClient, gpsPath string, runConfig TTNConfig) Manager {
	isGPS := gpsPath != ""
	statusMgr := NewStatusManager(ctx, netClient.FrequencyPlan(), runConfig.GatewayDescription, isGPS, netClient.DefaultLocation())

	bootTimeSetters := NewMultipleBootTimeSetter()
	bootTimeSetters.Add(statusMgr)

	return Manager{
		ctx:             ctx,
		conf:            conf,
		netClient:       netClient,
		statusMgr:       statusMgr,
		bootTimeSetters: bootTimeSetters,
		isGPS:           isGPS,
		// At the beginning, until we get our first uplinks, we keep a high polling rate to the concentrator
		uplinkPollingRate:   initUplinkPollingRate,
		downlinksSendMargin: runConfig.DownlinksSendMargin,
		ignoreCRC:           runConfig.IgnoreCRC,
	}
}

func (m *Manager) run() error {
	runStart := time.Now()
	m.ctx.WithField("DateTime", runStart).Info("Starting concentrator...")
	err := wrapper.StartLoRaGateway()
	if err != nil {
		return err
	}

	m.ctx.WithField("DateTime", time.Now()).Info("Concentrator started, packets can now be received and sent")
	err = m.handler(runStart)
	if shutdownErr := m.shutdown(); shutdownErr != nil {
		m.ctx.WithError(shutdownErr).Error("Couldn't stop concentrator gracefully")
	}
	return err
}

func (m *Manager) handler(runStart time.Time) (err error) {
	// First, we'll handle the case when the user wants to end the program
	c := make(chan os.Signal)
	defer close(c)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGABRT)

	// We'll start the routines, and attach them a context
	bgCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	routinesErr := m.startRoutines(bgCtx, runStart)
	defer close(routinesErr)

	// Finally, we'll listen to the different issues
	select {
	case sig := <-c:
		m.ctx.WithField("Signal", sig.String()).Info("Stopping packet forwarder")
	case err = <-routinesErr:
		m.ctx.Error("Program ended after one of the network links failed")
	}

	return err
}

func (m *Manager) findConcentratorBootTime(packets []wrapper.Packet, runStart time.Time) error {
	currentTime := time.Now()
	highestTimestamp := uint32(0)
	for _, p := range packets {
		if p.CountUS > highestTimestamp {
			highestTimestamp = p.CountUS
		}
	}
	if highestTimestamp == 0 {
		return nil
	}

	// Estimated boot time: highest timestamp (closest to current time) substracted to the current time
	highestTimestampDuration := time.Duration(highestTimestamp) * time.Microsecond
	bootTime := currentTime.Add(-highestTimestampDuration)
	if runStart.After(bootTime) || bootTime.After(time.Now()) {
		// Absurd timestamp
		return errors.New("Absurd uptime received by concentrator")
	}
	m.ctx.WithField("BootTime", bootTime).Info("Determined concentrator boot time")
	m.setBootTime(bootTime)
	return nil
}

func (m *Manager) setBootTime(bootTime time.Time) {
	m.bootTimeSetters.SetBootTime(bootTime)
	m.foundBootTime = true
	m.uplinkPollingRate = stableUplinkPollingRate
}

func (m *Manager) uplinkRoutine(bgCtx context.Context, runStart time.Time) chan error {
	errC := make(chan error)
	go func() {
		m.ctx.Info("Waiting for uplink packets")
		defer close(errC)
		for {
			packets, err := wrapper.Receive()
			if err != nil {
				errC <- errors.Wrap(err, "Uplink packets retrieval error")
				return
			}
			if len(packets) == 0 { // Empty payload => we sleep, then reiterate.
				time.Sleep(m.uplinkPollingRate)
				continue
			}

			m.ctx.WithField("NbPackets", len(packets)).Info("Received uplink packets")
			if !m.foundBootTime {
				// First packets received => find concentrator boot time
				err = m.findConcentratorBootTime(packets, runStart)
				if err != nil {
					m.ctx.WithError(err).Warn("Error when computing concentrator boot time - using packet forwarder run start time")
					m.setBootTime(runStart)
				}
			}

			validPackets := wrapUplinkPayload(m.ctx, packets, m.ignoreCRC, m.netClient.GatewayID())
			m.statusMgr.HandledRXBatch(len(validPackets), len(packets))
			if len(validPackets) == 0 {
				// Packets received, but with invalid CRC - ignoring
				time.Sleep(m.uplinkPollingRate)
				continue
			}

			m.ctx.WithField("NbValidPackets", len(validPackets)).Info("Sending valid uplink packets")
			m.netClient.SendUplinks(validPackets)

			select {
			case <-bgCtx.Done():
				errC <- nil
				return
			default:
				continue
			}
		}
	}()
	return errC
}

func (m *Manager) gpsRoutine(bgCtx context.Context) chan error {
	errC := make(chan error)
	go func() {
		m.ctx.Info("Starting GPS update routine")
		defer close(errC)
		for {
			select {
			case <-bgCtx.Done():
				return
			default:
				// The GPS time reference and coordinates are updated at `gpsUpdateRate`
				err := wrapper.UpdateGPSData(m.ctx)
				if err != nil {
					errC <- errors.Wrap(err, "GPS update error")
				}
			}
		}
	}()
	return errC
}

func (m *Manager) downlinkRoutine(bgCtx context.Context) {
	m.ctx.Info("Waiting for downlink messages")
	downlinkQueue := m.netClient.Downlinks()
	dManager := NewDownlinkManager(bgCtx, m.ctx, m.conf, m.statusMgr, m.downlinksSendMargin)
	m.bootTimeSetters.Add(dManager)
	for {
		select {
		case downlink := <-downlinkQueue:
			m.ctx.Info("Scheduling newly-received downlink packet")
			m.statusMgr.ReceivedTX()
			dManager.ScheduleDownlink(downlink)
		case <-bgCtx.Done():
			return
		}
	}
}

func (m *Manager) statusRoutine(bgCtx context.Context) chan error {
	errC := make(chan error)
	go func() {
		defer close(errC)
		for {
			select {
			case <-time.After(statusRoutineSleepRate):
				rtt, err := m.netClient.Ping()
				m.ctx.WithField("RTT", rtt).Debug("Ping to the router successful")
				if err != nil {
					errC <- errors.Wrap(err, "Network server health check error")
					return
				}

				status, err := m.statusMgr.GenerateStatus(rtt)
				if err != nil {
					errC <- errors.Wrap(err, "Gateway status computation error")
					return
				}

				err = m.netClient.SendStatus(*status)
				if err != nil {
					errC <- errors.Wrap(err, "Gateway status transmission error")
					return
				}
			case <-bgCtx.Done():
				return
			}
		}
	}()
	return errC
}

func (m *Manager) networkRoutine(bgCtx context.Context) chan error {
	errC := make(chan error)
	go func() {
		defer close(errC)
		if err := m.netClient.RefreshRoutine(bgCtx); err != nil {
			errC <- errors.Wrap(err, "Couldn't refresh account server token")
		}
	}()
	return errC
}

func (m *Manager) startRoutines(bgCtx context.Context, runTime time.Time) chan error {
	err := make(chan error)
	go func() {
		upCtx, upCancel := context.WithCancel(bgCtx)
		downCtx, downCancel := context.WithCancel(bgCtx)
		statusCtx, statusCancel := context.WithCancel(bgCtx)
		gpsCtx, gpsCancel := context.WithCancel(bgCtx)
		networkCtx, networkCancel := context.WithCancel(bgCtx)

		go m.downlinkRoutine(downCtx)
		uplinkErrors := m.uplinkRoutine(upCtx, runTime)
		statusErrors := m.statusRoutine(statusCtx)
		networkErrors := m.networkRoutine(networkCtx)
		var gpsErrors chan error
		if m.isGPS {
			gpsErrors = m.gpsRoutine(gpsCtx)
		}
		select {
		case uplinkError := <-uplinkErrors:
			err <- errors.Wrap(uplinkError, "Uplink routine error")
		case statusError := <-statusErrors:
			err <- errors.Wrap(statusError, "Status routine error")
		case networkError := <-networkErrors:
			err <- errors.Wrap(networkError, "Network routine error")
		case gpsError := <-gpsErrors:
			err <- errors.Wrap(gpsError, "GPS routine error")
		case <-bgCtx.Done():
			err <- nil
		}
		upCancel()
		gpsCancel()
		downCancel()
		statusCancel()
		networkCancel()
	}()
	return err
}

func (m *Manager) shutdown() error {
	m.netClient.Stop()
	return stopGateway(m.ctx)
}

func stopGateway(ctx log.Interface) error {
	err := wrapper.StopLoRaGateway()
	if err != nil {
		return err
	}

	ctx.Info("Concentrator stopped gracefully")
	return nil
}
