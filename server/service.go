package server

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/smancke/guble/protocol"
	"github.com/smancke/guble/server/webserver"

	"net/http"
	"reflect"
	"time"

	"github.com/docker/distribution/health"
	"github.com/smancke/guble/metrics"
)

const (
	defaultHealthFrequency = time.Second * 60
	defaultHealthThreshold = 1
)

var loggerService = log.WithField("module", "service")

// Service is the main struct for controlling a guble server
type Service struct {
	webserver       *webserver.WebServer
	router          Router
	modules         []module
	healthEndpoint  string
	healthFrequency time.Duration
	healthThreshold int
	metricsEndpoint string
}

// NewService creates a new Service, using the given Router and WebServer.
// If the router has already a configured Cluster, it is registered as a service module.
// The Router and Webserver are then registered as modules.
func NewService(router Router, webserver *webserver.WebServer) *Service {
	s := &Service{
		webserver:       webserver,
		router:          router,
		healthFrequency: defaultHealthFrequency,
		healthThreshold: defaultHealthThreshold,
	}
	cluster := router.Cluster()
	if cluster != nil {
		s.RegisterModules(1, 5, cluster)
		router.Cluster().MessageHandler = router
	}
	s.RegisterModules(2, 2, s.router)
	s.RegisterModules(3, 4, s.webserver)
	return s
}

// RegisterModules adds more modules (which can be Startable, Stopable, Endpoint etc.) to the service,
// with their start and stop ordering across all the service's modules.
func (s *Service) RegisterModules(startOrder int, stopOrder int, ifaces ...interface{}) {
	loggerService.WithFields(log.Fields{
		"numberOfNewModules":      len(ifaces),
		"numberOfExistingModules": len(s.modules),
	}).Info("RegisterModules")

	for _, i := range ifaces {
		m := module{
			iface:      i,
			startLevel: startOrder,
			stopLevel:  stopOrder,
		}
		s.modules = append(s.modules, m)
	}
}

// HealthEndpoint sets the endpoint used for health. Parameter for disabling the endpoint is: "". Returns the updated service.
func (s *Service) HealthEndpoint(endpointPrefix string) *Service {
	s.healthEndpoint = endpointPrefix
	return s
}

// MetricsEndpoint sets the endpoint used for metrics. Parameter for disabling the endpoint is: "". Returns the updated service.
func (s *Service) MetricsEndpoint(endpointPrefix string) *Service {
	s.metricsEndpoint = endpointPrefix
	return s
}

// Start checks the modules for the following interfaces and registers and/or starts:
//   Startable:
//   health.Checker:
//   Endpoint: Register the handler function of the Endpoint in the http service at prefix
func (s *Service) Start() error {
	el := protocol.NewErrorList("service: errors occured while starting: ")
	if s.healthEndpoint != "" {
		loggerService.WithField("healthEndpoint", s.healthEndpoint).Info("Health endpoint")
		s.webserver.Handle(s.healthEndpoint, http.HandlerFunc(health.StatusHandler))
	} else {
		loggerService.Info("Health endpoint disabled")
	}
	if s.metricsEndpoint != "" {
		loggerService.WithField("metricsEndpoint", s.metricsEndpoint).Info("Metrics endpoint")
		s.webserver.Handle(s.metricsEndpoint, http.HandlerFunc(metrics.HttpHandler))
	} else {
		loggerService.Info("Metrics endpoint disabled")
	}
	for order, iface := range s.ModulesSortedByStartOrder() {
		name := reflect.TypeOf(iface).String()
		if startable, ok := iface.(startable); ok {
			loggerService.WithFields(log.Fields{"name": name, "order": order}).Info("Starting module")
			if err := startable.Start(); err != nil {
				loggerService.WithError(err).WithField("name", name).Error("Error while starting module")
				el.Add(err)
			}
		} else {
			loggerService.WithFields(log.Fields{"name": name, "order": order}).Debug("Module is not startable")
		}
		if checker, ok := iface.(health.Checker); ok && s.healthEndpoint != "" {
			loggerService.WithField("name", name).Info("Registering module as Health-Checker")
			health.RegisterPeriodicThresholdFunc(name, s.healthFrequency, s.healthThreshold, health.CheckFunc(checker.Check))
		}
		if endpoint, ok := iface.(endpoint); ok {
			prefix := endpoint.GetPrefix()
			loggerService.WithFields(log.Fields{"name": name, "prefix": prefix}).Info("Registering module as Endpoint")
			s.webserver.Handle(prefix, endpoint)
		}
	}
	return el.ErrorOrNil()
}

// Stop stops the registered modules in their given order
func (s *Service) Stop() error {
	errors := protocol.NewErrorList("service stopping errors: ")
	for order, iface := range s.modulesSortedBy(ascendingStopOrder) {
		name := reflect.TypeOf(iface).String()
		if stoppable, ok := iface.(stopable); ok {
			loggerService.WithFields(log.Fields{"name": name, "order": order}).Info("Stopping module")
			if err := stoppable.Stop(); err != nil {
				errors.Add(err)
			}
		} else {
			loggerService.WithFields(log.Fields{"name": name, "order": order}).Debug("Module is not stoppable")
		}
	}
	if err := errors.ErrorOrNil(); err != nil {
		return fmt.Errorf("service: errors while stopping modules: %s", err)
	}
	return nil
}

// WebServer returns the service *webserver.WebServer instance
func (s *Service) WebServer() *webserver.WebServer {
	return s.webserver
}

// ModulesSortedByStartOrder returns the registered modules sorted by their startOrder property
func (s *Service) ModulesSortedByStartOrder() []interface{} {
	return s.modulesSortedBy(ascendingStartOrder)
}

// modulesSortedBy returns the registered modules sorted using a `by` criteria.
func (s *Service) modulesSortedBy(criteria by) []interface{} {
	var sorted []interface{}
	by(criteria).sort(s.modules)
	for _, m := range s.modules {
		sorted = append(sorted, m.iface)
	}
	return sorted
}
