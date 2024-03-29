// Copyright © 2017 The Things Network. Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package pktfwd

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/fields"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/health"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	tokenRefreshMargin = -2 * time.Minute
	uplinksBufferSize  = 32
)

type TTNConfig struct {
	ID                  string
	Key                 string
	AuthServer          string
	DiscoveryServer     string
	Router              string
	Version             string
	GatewayDescription  string
	DownlinksSendMargin time.Duration
	IgnoreCRC           bool
}

type TTNClient struct {
	antennaLocation *account.AntennaLocation
	routerConn      *grpc.ClientConn
	ctx             log.Interface
	uplinkStream    router.UplinkStream
	uplinkMutex     sync.Mutex
	downlinkStream  router.DownlinkStream
	statusStream    router.GatewayStatusStream
	account         *account.Account
	runConfig       TTNConfig
	connected       bool
	networkMutex    *sync.Mutex
	streamsMutex    *sync.Mutex
	token           string
	tokenExpiry     time.Time
	frequencyPlan   string
	// Communication between internal goroutines
	stopDownlinkQueue          chan bool
	stopUplinkQueue            chan bool
	stopMainRouterReconnection chan bool
	downlinkStreamChange       chan bool
	downlinkQueue              chan *router.DownlinkMessage
	uplinkQueue                chan *router.UplinkMessage
	routerChanges              chan func(c *TTNClient) error
}

type NetworkClient interface {
	SendStatus(status gateway.Status) error
	SendUplinks(messages []router.UplinkMessage)
	FrequencyPlan() string
	Downlinks() <-chan *router.DownlinkMessage
	GatewayID() string
	Ping() (time.Duration, error)
	DefaultLocation() *account.AntennaLocation
	Stop()
	RefreshRoutine(ctx context.Context) error
}

func (c *TTNClient) GatewayID() string {
	return c.runConfig.ID
}

type RouterHealthCheck struct {
	Conn     *grpc.ClientConn
	Duration time.Duration
	Err      error
}

func connectionHealthCheck(conn *grpc.ClientConn) (time.Duration, error) {
	timeBefore := time.Now()
	ok, err := health.Check(conn)
	if !ok {
		err = errors.New("Health check with the router failed")
	}
	return time.Now().Sub(timeBefore), err
}

func connectToRouter(ctx log.Interface, discoveryClient discovery.Client, router string) (*grpc.ClientConn, error) {
	routerAccess, err := discoveryClient.Get("router", router)
	if err != nil {
		return nil, err
	}

	var announcement = *routerAccess

	ctx.WithField("RouterID", router).Info("Connecting to router...")
	return announcement.Dial()
}

func (c *TTNClient) DefaultLocation() *account.AntennaLocation {
	return c.antennaLocation
}

func (c *TTNClient) FrequencyPlan() string {
	return c.frequencyPlan
}

func reconnectionDelay(tries uint) time.Duration {
	return time.Duration(math.Exp(float64(tries)/2.0)) * time.Second
}

func (c *TTNClient) tryMainRouterReconnection(gw account.Gateway, discoveryClient discovery.Client) {
	tries := uint(0)
	for {
		select {
		case <-c.stopMainRouterReconnection:
			return
		case <-time.After(reconnectionDelay(tries)):
			break
		}
		c.ctx.Info("Trying to reconnect to main router")
		routerConn, err := connectToRouter(c.ctx, discoveryClient, gw.Router.ID)
		if err != nil {
			c.ctx.WithError(err).Warn("Couldn't connect to the main router")
			tries = tries + 1
			continue
		}

		c.routerChanges <- func(t *TTNClient) error {
			t.routerConn = routerConn
			return nil
		}
		c.ctx.Info("Connection to main router successful")
		break
	}
}

func (c *TTNClient) Ping() (time.Duration, error) {
	c.networkMutex.Lock()
	defer c.networkMutex.Unlock()
	t, err := connectionHealthCheck(c.routerConn)
	return t, err
}

func (c *TTNClient) getLowestLatencyRouter(discoveryClient discovery.Client, fallbackRouters []account.GatewayRouter) (*grpc.ClientConn, error) {
	routerAnnouncements := make([]*discovery.Announcement, 0)
	for _, router := range fallbackRouters {
		routerAnnouncement, err := discoveryClient.Get("router", router.ID)
		if err != nil {
			continue
		}
		routerAnnouncements = append(routerAnnouncements, routerAnnouncement)
	}
	return c.getLowestLatencyRouterFromAnnouncements(discoveryClient, routerAnnouncements)
}

func (c *TTNClient) getLowestLatencyRouterFromAnnouncements(discoveryClient discovery.Client, routerAnnouncements []*discovery.Announcement) (*grpc.ClientConn, error) {
	var routerConn *grpc.ClientConn
	routerHealthChannel := make(chan RouterHealthCheck)
	for _, routerAnnouncement := range routerAnnouncements {
		announcement := routerAnnouncement
		go func() {
			conn, err := announcement.Dial()
			if err != nil {
				routerHealthChannel <- RouterHealthCheck{Err: err}
				return
			}
			duration, err := connectionHealthCheck(conn)
			routerHealthChannel <- RouterHealthCheck{
				Err:      err,
				Duration: duration,
				Conn:     conn,
			}
		}()
	}

	lowestPing := time.Duration(math.MaxInt64)
	routersChecked := 0
	for routerHealth := range routerHealthChannel {
		if routerHealth.Err == nil && routerHealth.Duration < lowestPing {
			if routerConn != nil {
				routerConn.Close()
			}
			routerConn = routerHealth.Conn
		}
		routersChecked++
		if routersChecked == len(routerAnnouncements) {
			break
		}
	}
	if routerConn == nil {
		return nil, errors.New("Packet forwarder couldn't establish a healthy connection with any router")
	}
	c.ctx.Info("Identified the lowest latency router")
	return routerConn, nil
}

func (c *TTNClient) getRouterClient(ctx log.Interface) error {
	ctx.WithField("Address", c.runConfig.DiscoveryServer).Info("Connecting to TTN discovery server")
	discoveryClient, err := discovery.NewClient(c.runConfig.DiscoveryServer, &discovery.Announcement{
		ServiceName:    "ttn-packet-forwarder",
		ServiceVersion: c.runConfig.Version,
		Id:             c.runConfig.ID,
	}, func() string { return "" })
	if err != nil {
		return err
	}
	ctx.Info("Connected to discovery server - getting router address")

	defer discoveryClient.Close()

	var routerConn *grpc.ClientConn
	if c.runConfig.Router == "" {
		gw, err := c.account.FindGateway(c.GatewayID())
		if err != nil {
			return errors.Wrap(err, "Couldn't fetch the gateway information from the account server")
		}

		if gw.Router.ID != "" {
			routerConn, err = connectToRouter(c.ctx.WithField("RouterID", gw.Router.ID), discoveryClient, gw.Router.ID)
		}
		if gw.Router.ID == "" || err != nil {
			if err != nil {
				ctx.WithError(err).WithField("RouterID", gw.Router.ID).Warn("Couldn't connect to main router - trying to connect to fallback routers")
			}
			fallbackRouters := gw.FallbackRouters
			if len(fallbackRouters) == 0 {
				ctx.Warn("No fallback routers in memory for this gateway - loading all routers")
				routers, err := discoveryClient.GetAll("router")
				if err != nil {
					ctx.WithError(err).Error("Couldn't retrieve routers")
					return err
				}
				routerConn, err = c.getLowestLatencyRouterFromAnnouncements(discoveryClient, routers)
				if err != nil {
					return errors.Wrap(err, "Couldn't figure out the lowest latency router")
				}
			} else {
				routerConn, err = c.getLowestLatencyRouter(discoveryClient, fallbackRouters)
				if err != nil {
					return errors.Wrap(err, "Couldn't figure out the lowest latency router")
				}
			}
			defer func() {
				// Wait for the function to be finished, to protect `c.routerConn`
				go c.tryMainRouterReconnection(gw, discoveryClient)
			}()
		}
	} else {
		routerConn, err = connectToRouter(ctx, discoveryClient, c.runConfig.Router)
		if err != nil {
			return errors.Wrap(err, "Couldn't connect to user-specified router")
		}
		ctx.Info("Connected to router")
	}

	c.routerConn = routerConn
	return nil
}

func (c *TTNClient) Downlinks() <-chan *router.DownlinkMessage {
	return c.downlinkQueue
}

func (c *TTNClient) queueUplinks() {
	for {
		select {
		case <-c.stopUplinkQueue:
			c.ctx.Info("Closing uplinks queue")
			close(c.uplinkQueue)
			return
		case uplink := <-c.uplinkQueue:
			ctx := c.ctx.WithFields(fields.Get(uplink))
			if err := c.uplinkStream.Send(uplink); err != nil {
				ctx.WithError(err).Warn("Uplink message transmission to the back-end failed.")
			} else {
				ctx.Info("Uplink message transmission successful.")
			}
		}
	}
}

func (c *TTNClient) queueDownlinks() {
	c.ctx.Info("Downlinks queuing routine started")
	c.streamsMutex.Lock()
	downlinkStreamChannel := c.downlinkStream.Channel()
	c.streamsMutex.Unlock()
	for {
		select {
		case <-c.stopDownlinkQueue:
			c.ctx.Info("Closing downlinks queue")
			close(c.downlinkQueue)
			return
		case downlink := <-downlinkStreamChannel:
			c.ctx.Info("Received downlink packet")
			c.downlinkQueue <- downlink
		case <-c.downlinkStreamChange:
			c.streamsMutex.Lock()
			downlinkStreamChannel = c.downlinkStream.Channel()
			c.streamsMutex.Unlock()
		}
	}
}

func (c *TTNClient) fetchAccountServerInfo() error {
	c.account = account.NewWithKey(c.runConfig.AuthServer, c.runConfig.Key)
	gw, err := c.account.FindGateway(c.runConfig.ID)
	if err != nil {
		return errors.Wrap(err, "Account server error")
	}
	c.antennaLocation = gw.AntennaLocation
	c.token = gw.Token.AccessToken
	c.tokenExpiry = gw.Token.Expiry
	c.frequencyPlan = gw.FrequencyPlan
	c.ctx.WithField("TokenExpiry", c.tokenExpiry).Info("Refreshed account server information")
	return nil
}

func (c *TTNClient) RefreshRoutine(ctx context.Context) error {
	for {
		refreshTime := c.tokenExpiry.Add(tokenRefreshMargin)
		c.ctx.Debugf("Preparing to update network clients at %v", refreshTime)
		select {
		case <-time.After(refreshTime.Sub(time.Now())):
			c.routerChanges <- func(t *TTNClient) error {
				if err := t.fetchAccountServerInfo(); err != nil {
					return errors.Wrap(err, "Couldn't update account server info")
				}
				return nil
			}
			c.ctx.Debug("Refreshed network connection")
		case <-ctx.Done():
			return nil
		}
	}
}

func CreateNetworkClient(ctx log.Interface, ttnConfig TTNConfig) (NetworkClient, error) {
	var client = &TTNClient{
		ctx:                  ctx,
		runConfig:            ttnConfig,
		downlinkQueue:        make(chan *router.DownlinkMessage),
		uplinkQueue:          make(chan *router.UplinkMessage, uplinksBufferSize),
		networkMutex:         &sync.Mutex{},
		streamsMutex:         &sync.Mutex{},
		stopDownlinkQueue:    make(chan bool),
		stopUplinkQueue:      make(chan bool),
		downlinkStreamChange: make(chan bool),
		routerChanges:        make(chan func(c *TTNClient) error),
	}

	client.networkMutex.Lock()
	defer client.networkMutex.Unlock()

	// Get the first token
	err := client.fetchAccountServerInfo()
	if err != nil {
		return nil, err
	}

	// Updating with the initial RouterConn
	err = client.getRouterClient(ctx)
	if err != nil {
		return nil, err
	}

	client.connectToStreams(router.NewRouterClientForGateway(router.NewRouterClient(client.routerConn), client.runConfig.ID, client.token))

	go client.watchRouterChanges()

	go client.queueDownlinks()
	go client.queueUplinks()

	return client, nil
}

func (c *TTNClient) watchRouterChanges() {
	for {
		select {
		case routerChange := <-c.routerChanges:
			if routerChange == nil { // Channel closed, shutting network client down
				return
			}
			c.networkMutex.Lock()
			if err := routerChange(c); err != nil {
				c.ctx.WithError(err).Warn("Couldn't operate network client change")
			} else {
				c.connectToStreams(router.NewRouterClientForGateway(router.NewRouterClient(c.routerConn), c.runConfig.ID, c.token))
				c.downlinkStreamChange <- true
			}
			c.networkMutex.Unlock()
		}
	}
}

func (c *TTNClient) connectToStreams(routerClient router.RouterClientForGateway) {
	c.streamsMutex.Lock()
	defer c.streamsMutex.Unlock()
	if c.connected {
		c.disconnectOfStreams()
	}
	c.uplinkStream = router.NewMonitoredUplinkStream(routerClient)
	c.downlinkStream = router.NewMonitoredDownlinkStream(routerClient)
	c.statusStream = router.NewMonitoredGatewayStatusStream(routerClient)
	c.connected = true
}

func (c *TTNClient) disconnectOfStreams() {
	c.uplinkStream.Close()
	c.downlinkStream.Close()
	c.statusStream.Close()
	c.connected = false
}

func (c *TTNClient) SendUplinks(messages []router.UplinkMessage) {
	for _, message := range messages {
		c.uplinkQueue <- &message
	}
}

func (c *TTNClient) SendStatus(status gateway.Status) error {
	var uptimeString string
	status.Region = c.frequencyPlan
	uptimeDuration, err := time.ParseDuration(fmt.Sprintf("%dus", status.GetTimestamp()))
	if err == nil {
		uptimeString = uptimeDuration.String()
	} else {
		uptimeString = fmt.Sprintf("%fs", float32(status.GetTimestamp())/1000000.0)
	}
	c.ctx.WithFields(log.Fields{
		"TXPacketsReceived": status.GetTxIn(),
		"TXPacketsValid":    status.GetTxOk(),
		"RXPacketsReceived": status.GetRxIn(),
		"RXPacketsValid":    status.GetRxOk(),
		"FrequencyPlan":     status.GetRegion(),
		"Uptime":            uptimeString,
		"Load1":             status.GetOs().GetLoad_1(),
		"Load5":             status.GetOs().GetLoad_5(),
		"Load15":            status.GetOs().GetLoad_15(),
		"CpuPercentage":     status.GetOs().GetCpuPercentage(),
		"MemoryPercentage":  status.GetOs().GetMemoryPercentage(),
		"Latitude":          status.GetGps().GetLatitude(),
		"Longitude":         status.GetGps().GetLongitude(),
		"Altitude":          status.GetGps().GetAltitude(),
		"RTT":               status.GetRtt(),
	}).Info("Sending status to the network server")
	err = c.statusStream.Send(&status)
	if err != nil {
		return errors.Wrap(err, "Status stream error")
	}
	return nil
}

// Stop a running network client
func (c *TTNClient) Stop() {
	c.stopDownlinkQueue <- true
	c.stopUplinkQueue <- true
	select {
	case c.stopMainRouterReconnection <- true:
		break
	default:
		break
	}
	close(c.routerChanges)
	c.streamsMutex.Lock()
	defer c.streamsMutex.Unlock()
	c.disconnectOfStreams()
}
