package sms

import (
	"context"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/cosminrentea/gobbler/protocol"
	"github.com/cosminrentea/gobbler/server/connector"
	"github.com/cosminrentea/expvarmetrics"
	"github.com/cosminrentea/gobbler/server/router"
	"github.com/cosminrentea/gobbler/server/store"
	"time"
)

const (
	SMSSchema       = "sms_notifications"
	SMSDefaultTopic = "/sms"
)

type Sender interface {
	Send(*protocol.Message) error
}

type Config struct {
	Enabled         *bool
	APIKey          *string
	APISecret       *string
	Workers         *int
	SMSTopic        *string
	IntervalMetrics *bool
	Toggleable      *bool

	Name   string
	Schema string
}

type gateway struct {
	config *Config

	sender Sender
	router router.Router
	route  *router.Route

	LastIDSent uint64

	ctx        context.Context
	cancelFunc context.CancelFunc

	logger *log.Entry
}

func New(router router.Router, sender Sender, config Config) (*gateway, error) {
	if *config.Workers <= 0 {
		*config.Workers = connector.DefaultWorkers
	}
	logger.WithField("number", *config.Workers).Info("sms workers")
	config.Schema = SMSSchema
	config.Name = SMSDefaultTopic
	return &gateway{
		config: &config,
		router: router,
		sender: sender,
		logger: logger.WithField("name", config.Name),
	}, nil
}

// Start the sms gateway; it is an idempotent operation.
func (g *gateway) Start() error {
	g.logger.Info("Starting gateway")
	if g.cancelFunc != nil {
		g.logger.Info("Gateway was already started")
		return nil
	}

	err := g.ReadLastID()
	if err != nil {
		g.logger.Error("Could not ReadLastID in Start")
		return err
	}

	g.ctx, g.cancelFunc = context.WithCancel(context.Background())
	g.initRoute(false)

	go g.Run()

	g.startMetrics()

	g.logger.Info("Started gateway")
	return nil
}

func (g *gateway) initRoute(fetch bool) {
	g.route = router.NewRoute(router.RouteConfig{
		Path:        protocol.Path(*g.config.SMSTopic),
		ChannelSize: 5000,
		QueueSize:   -1,
		Timeout:     -1,
	})
	if fetch || !*g.config.Toggleable {
		g.route.FetchRequest = g.fetchRequest()
	}
}

func (g *gateway) fetchRequest() (fr *store.FetchRequest) {
	if g.LastIDSent > 0 {
		fr = store.NewFetchRequest(
			protocol.Path(*g.config.SMSTopic).Partition(),
			g.LastIDSent+1,
			0,
			store.DirectionForward,
			-1)
	}
	return
}

func (g *gateway) Run() {
	g.logger.Info("Run gateway")
	var provideErr error
	go func() {
		err := g.route.Provide(g.router, true)
		if err != nil {
			// cancel subscription loop if there is an error on the provider
			logger.WithError(err).Error("Provide returned error")
			provideErr = err
			g.Cancel()
		} else {
			g.logger.Debug("Provide ok")
		}
	}()

	err := g.proxyLoop()
	if err != nil && provideErr == nil {
		// If Route channel closed, try restarting
		if err == connector.ErrRouteChannelClosed {
			g.logger.WithError(err).Info("Restarting")
			g.restart()
			return
		}
	}

	if provideErr != nil {
		// TODO Bogdan Treat errors where a subscription provide fails
		g.logger.WithField("error", provideErr.Error()).Error("Route provide error")

		// Router closed the route, try restart
		if provideErr == router.ErrInvalidRoute {
			g.logger.WithError(provideErr).Info("Restarting")
			g.restart()
			return
		}
		// Router module is stopping, exit the process
		if _, ok := provideErr.(*router.ModuleStoppingError); ok {
			g.logger.Info("SMS Gateway is exiting.Router is stopping")
			return
		}
	}
}

func (g *gateway) proxyLoop() error {
	var (
		opened      bool = true
		receivedMsg *protocol.Message
	)
	defer func() { g.cancelFunc = nil }()

	for opened {
		select {
		case receivedMsg, opened = <-g.route.MessagesChannel():
			if !opened {
				logger.WithField("receivedMsg", receivedMsg).Info("not open")
				break
			}

			err := g.send(receivedMsg)
			if err == ErrRetryFailed || err == ErrLastIDCouldNotBeSet || err == ErrSmsTooLong {
				// THIS MAY BE BLOCKING.Maybe not a good idea.
				for errSetLastSentId := g.SetLastSentID(receivedMsg.ID); errSetLastSentId != nil; {
					g.logger.WithError(errSetLastSentId).Error("Error setting last ID, retrying")
					time.Sleep(time.Second)
				}
				g.logger.WithError(err).WithField("id", receivedMsg.ID).Info("Set last sent id after acceptable error occurred")
				continue
			} else if err != nil {
				g.logger.WithError(err).Error("Exiting from proxyLoop.")
				return err
			}
		case <-g.ctx.Done():
			return nil
		}
	}

	//TODO Cosmin Bogdan returning this error can mean 2 things: overflow of route's channel, or intentional stopping of router / gubled.
	return connector.ErrRouteChannelClosed
}

func (g *gateway) send(receivedMsg *protocol.Message) error {
	err := g.sender.Send(receivedMsg)
	if err != nil {
		log.WithField("error", err.Error()).Error("Sending of message failed")
		mTotalResponseErrors.Add(1)
		pNexmoResponseErrors.Inc()
		return err
	}
	mTotalSentMessages.Add(1)
	pSent.Inc()
	err = g.SetLastSentID(receivedMsg.ID)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Setting last id for sms connector failed.")
		return ErrLastIDCouldNotBeSet
	}

	return nil
}

func (g *gateway) restart() error {
	g.logger.WithField("LastIDSent", g.LastIDSent).Info("SMS Gateway restarting")

	g.Cancel()
	g.cancelFunc = nil

	err := g.ReadLastID()
	if err != nil {
		g.logger.WithError(err).Error("Could not ReadLastID in restart")
		return err
	}

	g.initRoute(true)

	go g.Run()

	g.logger.WithField("LastIDSent", g.LastIDSent).Info("SMS Gateway restarted")
	return nil
}

// Stop the sms gateway; it is an idempotent operation.
func (g *gateway) Stop() error {
	g.logger.Info("Stopping gateway")
	if g.cancelFunc == nil {
		g.logger.Info("Gateway was already stopped")
		return nil
	}
	if *g.config.Toggleable {
		g.logger.Info("Unsubscribing the sms route")
		g.router.Unsubscribe(g.route)
	}
	g.logger.Info("Calling the cancel function")
	g.cancelFunc()
	g.cancelFunc = nil
	g.logger.Info("Stopped gateway")

	return nil
}

func (g *gateway) SetLastSentID(ID uint64) error {
	g.logger.WithField("LastIDSent", ID).WithField("path", *g.config.SMSTopic).Debug("Seting LastIDSent")

	kvStore, err := g.router.KVStore()
	if err != nil {
		g.logger.WithField("error", err.Error()).Error("KVStore could not be accesed from gateway")
		return err
	}

	data, err := json.Marshal(struct{ ID uint64 }{ID: ID})
	if err != nil {
		g.logger.WithField("error", err.Error()).Error("Error encoding last ID")
		return err
	}
	err = kvStore.Put(g.config.Schema, *g.config.SMSTopic, data)
	if err != nil {
		g.logger.WithField("error", err.Error()).WithField("path", *g.config.SMSTopic).Error("KVStore could not set value for LastIDSent for topic")
		return err
	}
	g.LastIDSent = ID
	return nil
}

func (g *gateway) ReadLastID() error {
	kvStore, err := g.router.KVStore()
	if err != nil {
		g.logger.WithField("error", err.Error()).Error("KVStore could not be accesed from sms gateway")
		return err
	}
	data, exist, err := kvStore.Get(g.config.Schema, *g.config.SMSTopic)
	if err != nil {
		g.logger.WithField("error", err.Error()).WithField("path", *g.config.SMSTopic).Error("KvStore could not get value for LastIDSent for topic")
		return err
	}
	if !exist {
		g.logger.Error("Setting LastIDSent to 0")
		g.LastIDSent = 0
		return nil
	}

	v := &struct{ ID uint64 }{}
	err = json.Unmarshal(data, v)
	if err != nil {
		g.logger.WithField("error", err.Error()).Error("Could not parse as uint64 the LastIDSent value stored in db")
		return err
	}
	g.LastIDSent = v.ID

	g.logger.WithField("LastIDSent", g.LastIDSent).WithField("path", *g.config.SMSTopic).Debug("ReadLastID")
	return nil
}

func (g *gateway) Cancel() {
	if g.cancelFunc != nil {
		g.cancelFunc()
	}
}

func (g *gateway) startMetrics() {
	mTotalSentMessages.Set(0)
	mTotalSendErrors.Set(0)
	mTotalResponseErrors.Set(0)
	mTotalResponseInternalErrors.Set(0)

	if *g.config.IntervalMetrics {
		g.startIntervalMetric(mMinute, time.Minute)
		g.startIntervalMetric(mHour, time.Hour)
		g.startIntervalMetric(mDay, time.Hour*24)
	}
}

func (g *gateway) startIntervalMetric(m metrics.Map, td time.Duration) {
	metrics.RegisterInterval(g.ctx, m, td, resetIntervalMetrics, processAndResetIntervalMetrics)
}
