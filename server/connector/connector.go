package connector

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/smancke/guble/server/router"
	"github.com/smancke/guble/server/service"
)

type Sender interface {
	// Send takes a Request and returns the response or error
	Send(Request) (interface{}, error)
}

type ResponseHandler interface {
	// HandleResponse handles the response+error returned by the Sender
	HandleResponse(Request, interface{}, error) error
}

type Connector interface {
	service.Startable
	service.Stopable
	service.Endpoint
}

type connector struct {
	config  ConnectorConfig
	sender  Sender
	manager Manager
	router  router.Router
	ctx     context.Context
}

type ConnectorConfig struct {
	Name   string
	Prefix string
	Url    string
}

func NewConnector(router router.Router, sender Sender, config ConnectorConfig) (Connector, error) {
	kvstore, err := router.KVStore()
	if err != nil {
		return nil, err
	}

	manager, err := NewManager(config.Name, kvstore)
	if err != nil {
		return nil, err
	}

	return &connector{
		manager: manager,
		config:  config,
		router:  router,
		sender:  sender,
		ctx:     context.Background(),
	}, nil
}

func (c *connector) GetPrefix() string {
	return c.config.Prefix
}

// TODO Bogdan Refactor this so the router is built one time
func (c *connector) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r := mux.NewRouter()

	base := r.PathPrefix(c.GetPrefix()).Subrouter()
	base.Methods("GET").HandlerFunc(c.GetList)

	s := base.Path(c.config.Url).Subrouter()
	s.Methods("POST").HandlerFunc(c.Post)
	s.Methods("DELETE").HandlerFunc(c.Delete)

	r.ServeHTTP(w, req)
}

// GetList returns list of subscribers
func (c *connector) GetList(w http.ResponseWriter, req *http.Request) {

}

// Post creates a new subscriber
func (c *connector) Post(w http.ResponseWriter, req *http.Request) {
	// vars := mux.Vars(req)
}

// Delete removes a subscriber
func (c *connector) Delete(w http.ResponseWriter, req *http.Request) {

}

// Start will run start all current subscriptions and workers to process the messages
func (c *connector) Start() error {
	return nil
}

// Stop stops the context
func (c *connector) Stop() error {
	return nil
}