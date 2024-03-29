// Copyright © 2017 The Things Network. Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package pktfwd

import (
	"context"
	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/queue"
	"github.com/TheThingsNetwork/packet_forwarder/util"
	"github.com/TheThingsNetwork/packet_forwarder/wrapper"
	"github.com/TheThingsNetwork/ttn/api/router"
)

// BootTimeSetter is an interface that implements every type that needs receive boot time (mostly to
// determine uptime afterwards)
type BootTimeSetter interface {
	SetBootTime(t time.Time)
}

type multipleBootTimeSetter struct {
	list []BootTimeSetter
	t    *time.Time
}

func NewMultipleBootTimeSetter() multipleBootTimeSetter {
	return multipleBootTimeSetter{
		list: make([]BootTimeSetter, 0),
	}
}

func (b *multipleBootTimeSetter) SetBootTime(t time.Time) {
	for _, receiver := range b.list {
		receiver.SetBootTime(t)
	}
	b.t = &t
}
func (b *multipleBootTimeSetter) Add(t BootTimeSetter) {
	if b.t != nil {
		t.SetBootTime(*b.t)
	}
	b.list = append(b.list, t)
}

// DownlinkManager is an interface that starts scheduling every downlink that is given to it
type DownlinkManager interface {
	BootTimeSetter
	ScheduleDownlink(d *router.DownlinkMessage)
}

type downlinkManager struct {
	queue              queue.JIT
	ctx                log.Interface
	conf               util.Config
	bgCtx              context.Context
	statusMgr          StatusManager
	startupTime        time.Time
	downlinkSendMargin time.Duration
}

func (d *downlinkManager) getTimeMargin() time.Duration {
	return d.downlinkSendMargin
}

// NewDownlinkManager returns a new downlink manager that runs as long as the context doesn't close
func NewDownlinkManager(bgCtx context.Context, ctx log.Interface, conf util.Config, statusMgr StatusManager, sendingTimeMargin time.Duration) DownlinkManager {
	downlinkMgr := &downlinkManager{
		queue:              queue.NewJIT(),
		ctx:                ctx,
		conf:               conf,
		bgCtx:              bgCtx,
		statusMgr:          statusMgr,
		downlinkSendMargin: sendingTimeMargin,
	}
	ctx.WithField("SendingTimeMargin", sendingTimeMargin).Debug("Configured margin between downlink sent and concentrator processing")
	go downlinkMgr.handleDownlinks()
	return downlinkMgr
}

func (d *downlinkManager) SetBootTime(t time.Time) {
	d.startupTime = t
}

func (d *downlinkManager) handleDownlinks() {
	downlinks := d.nextDownlinks()
	for {
		select {
		case downlink := <-downlinks:
			d.ctx.WithField("ConcentratorUptime", time.Now().Sub(d.startupTime)).Info("Received downlink from JIT queue, transmitting to the concentrator")
			if err := wrapper.SendDownlink(downlink, d.conf, d.ctx); err == nil {
				d.statusMgr.SentTX()
			}
		case <-d.bgCtx.Done():
			d.ctx.Info("Stopping downlink manager")
			return
		}
	}
}

func (d *downlinkManager) nextDownlinks() chan *router.DownlinkMessage {
	downlink := make(chan *router.DownlinkMessage)
	go func() {
		for {
			item := d.queue.Next()
			if item == nil {
				d.ctx.Warn("JIT queue closing, no more downlinks sent")
				break
			}
			next := item.(*router.DownlinkMessage)
			select {
			case downlink <- next:
			default:
			}
		}
		close(downlink)
	}()

	return downlink
}

func (d *downlinkManager) ScheduleDownlink(message *router.DownlinkMessage) {
	lora := message.ProtocolConfiguration.GetLorawan()
	if lora == nil {
		d.ctx.Warn("Received non-LORA downlink, ignoring")
		return
	}

	margin := d.getTimeMargin()

	schedulingTimestamp := util.TXTimestamp(message.GetGatewayConfiguration().GetTimestamp())
	d.ctx.WithFields(log.Fields{
		"ExpectedSendingTimestamp": schedulingTimestamp.GetAsDuration(),
		"ConcentratorBootTime":     d.startupTime,
		"ConcentratorUptime":       time.Now().Sub(d.startupTime),
		"SchedulingTimestamp":      d.startupTime.Add(-margin).Add(schedulingTimestamp.GetAsDuration()),
	}).Info("Scheduled downlink")
	d.queue.Schedule(message, d.startupTime.Add(-margin).Add(schedulingTimestamp.GetAsDuration()))
}
