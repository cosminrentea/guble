package fcm

import (
	"fmt"

	"time"

	"github.com/Bogh/gcm"
	"github.com/cosminrentea/expvarmetrics"
	"github.com/cosminrentea/gobbler/protocol"
	"github.com/cosminrentea/gobbler/server/connector"
	"github.com/cosminrentea/gobbler/server/kafka"
	"github.com/cosminrentea/gobbler/server/router"
)

const (
	// schema is the default database schema for FCM
	schema = "fcm_registration"

	deviceTokenKey = "device_token"
	userIDKEy      = "user_id"
)

// Config is used for configuring the Firebase Cloud Messaging component.
type Config struct {
	Enabled              *bool
	APIKey               *string
	Workers              *int
	Endpoint             *string
	Prefix               *string
	IntervalMetrics      *bool
	AfterMessageDelivery protocol.MessageDeliveryCallback
}

// Connector is the structure for handling the communication with Firebase Cloud Messaging
type fcm struct {
	Config
	connector.Connector
	fcmKafkaReportingTopic string
}

// New creates a new *fcm and returns it as an connector.ResponsiveConnector
func New(router router.Router, sender connector.Sender, config Config, kafkaProducer kafka.Producer, kafkaReportingTopic string, fcmKafkaReportingTopic string) (connector.ResponsiveConnector, error) {
	baseConn, err := connector.NewConnector(router, sender, connector.Config{
		Name:       "fcm",
		Schema:     schema,
		Prefix:     *config.Prefix,
		URLPattern: fmt.Sprintf("/{%s}/{%s}/{%s:.*}", deviceTokenKey, userIDKEy, connector.TopicParam),
		Workers:    *config.Workers,
	},
		kafkaProducer,
		kafkaReportingTopic,
	)

	if err != nil {
		logger.WithError(err).Error("Base connector error")
		return nil, err
	}

	f := &fcm{config, baseConn,fcmKafkaReportingTopic}
	f.SetResponseHandler(f)
	return f, nil
}

func (f *fcm) Start() error {
	err := f.Connector.Start()
	if err == nil {
		f.startMetrics()
	}
	return err
}

func (f *fcm) Stop() error {
	return f.Connector.Stop()
}

func (f *fcm) startMetrics() {
	mTotalSentMessages.Set(0)
	mTotalSendErrors.Set(0)
	mTotalResponseErrors.Set(0)
	mTotalResponseInternalErrors.Set(0)
	mTotalResponseNotRegisteredErrors.Set(0)
	mTotalReplacedCanonicalErrors.Set(0)
	mTotalResponseOtherErrors.Set(0)

	if *f.IntervalMetrics {
		f.startIntervalMetric(mMinute, time.Minute)
		f.startIntervalMetric(mHour, time.Hour)
		f.startIntervalMetric(mDay, time.Hour*24)
	}
}

func (f *fcm) startIntervalMetric(m metrics.Map, td time.Duration) {
	metrics.RegisterInterval(f.Context(), m, td, resetIntervalMetrics, processAndResetIntervalMetrics)
}

func (f *fcm) HandleResponse(request connector.Request, responseIface interface{}, metadata *connector.Metadata, err error) error {
	l := logger.WithField("correlation_id", request.Message().CorrelationID())

	event := FcmEvent{
		Payload: kafka.PushEventPayload{},
	}

	errFill := event.fillApnsEvent(request)
	if errFill != nil {
		logger.WithError(errFill).Error("Error filling event")
	}


	if err != nil && !isValidResponseError(err) {
		l.WithField("error", err.Error()).Error("Error sending message to FCM")
		mTotalSendErrors.Add(1)
		pSendErrors.Inc()
		if *f.IntervalMetrics && metadata != nil {
			addToLatenciesAndCountsMaps(currentTotalErrorsLatenciesKey, currentTotalErrorsKey, metadata.Latency)
		}
		return err
	}
	message := request.Message()
	subscriber := request.Subscriber()

	response, ok := responseIface.(*gcm.Response)
	if !ok {
		mTotalResponseErrors.Add(1)
		pResponseErrors.Inc()
		return fmt.Errorf("Invalid FCM Response")
	}

	l.WithField("messageID", message.ID).Debug("Delivered message to FCM")

	subscriber.SetLastID(message.ID)
	if err := f.Manager().Update(request.Subscriber()); err != nil {
		l.WithField("error", err.Error()).Error("Manager could not update subscription")
		mTotalResponseInternalErrors.Add(1)
		return err
	}
	if response.Ok() {

		event.Payload.Status = "Success"
		event.Payload.ErrorText = ""

		mTotalSentMessages.Add(1)
		pSent.Inc()
		if *f.IntervalMetrics && metadata != nil {
			addToLatenciesAndCountsMaps(currentTotalMessagesLatenciesKey, currentTotalMessagesKey, metadata.Latency)
		}

		err := event.report(f.KafkaProducer(), f.fcmKafkaReportingTopic)
		if err !=nil && err != errFcmKafkaReportingConfiguration {
			logger.WithError(err).Error("Reporting APNS to kafka failed")
		}

		return nil
	}

	l.WithField("success", response.Success).Debug("Handling FCM Error")

	errText := response.Error.Error()
	event.Payload.ErrorText = errText
	event.Payload.Status = "Fail"

	switch errText {
	case "NotRegistered":
		l.Debug("Removing not registered FCM subscription")
		f.Manager().Remove(subscriber)
		mTotalResponseNotRegisteredErrors.Add(1)
		pResponseNotRegisteredErrors.Inc()
		return response.Error
	case "InvalidRegistration":
		l.WithField("jsonError", errText).Error("InvalidRegistration of FCM subscription")
	default:
		l.WithField("jsonError", errText).Error("Unexpected error while sending to FCM")
	}

	err = event.report(f.KafkaProducer(), f.fcmKafkaReportingTopic)
	if err !=nil && err != errFcmKafkaReportingConfiguration {
		logger.WithError(err).Error("Reporting APNS to kafka failed")
	}

	if response.CanonicalIDs != 0 {
		mTotalReplacedCanonicalErrors.Add(1)
		pResponseReplacedCanonicalErrors.Inc()
		// we only send to one receiver, so we know that we can replace the old id with the first registration id (=canonical id)
		return f.replaceCanonical(request.Subscriber(), response.Results[0].RegistrationID)
	}
	mTotalResponseOtherErrors.Add(1)
	pResponseOtherErrors.Inc()
	return nil
}

func (f *fcm) replaceCanonical(subscriber connector.Subscriber, newToken string) error {
	manager := f.Manager()
	err := manager.Remove(subscriber)
	if err != nil {
		return err
	}

	topic := subscriber.Route().Path
	params := subscriber.Route().RouteParams.Copy()

	params[deviceTokenKey] = newToken

	newSubscriber, err := manager.Create(topic, params)
	go f.Run(newSubscriber)
	return err
}
