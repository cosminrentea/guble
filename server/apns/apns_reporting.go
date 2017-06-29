package apns

import (
	"encoding/json"
	"errors"
	"time"

	"fmt"

	"github.com/cosminrentea/go-uuid"
	"github.com/cosminrentea/gobbler/server/connector"
	"github.com/cosminrentea/gobbler/server/kafka"
)

type ApnsEvent struct {
	ID      string                 `json:"id"`
	Time    string                 `json:"time"`
	Type    string                 `json:"type"`
	Payload kafka.PushEventPayload `json:"payload"`
}

var (
	errApnsKafkaReportingConfiguration = errors.New("Kafka Reporting for APNS is not correctly configured")
	errAlertDecodingFailed             = errors.New("Decoding of aps2.alert field failed")
)

func (ev *ApnsEvent) fillApnsEvent(request connector.Request) error {
	ev.Type = "push_notification_information"

	deviceID := request.Subscriber().Route().Get(deviceIDKey)
	ev.Payload.DeviceID = deviceID

	userID := request.Subscriber().Route().Get(userIDKey)
	ev.Payload.UserID = userID

	var payload Payload

	err := json.Unmarshal(request.Message().Body, &payload)
	if err != nil {
		logger.WithError(err).Error("Error reading apns notification built.")
		return err
	}

	ev.Payload.DeepLink = payload.Deeplink

	alert := payload.Aps.Alert
	alertBody, ok := alert.(map[string]interface{})
	if !ok {
		return errAlertDecodingFailed
	}
	ev.Payload.NotificationBody = fmt.Sprintf("%s", alertBody["body"])
	ev.Payload.NotificationTitle = fmt.Sprintf("%s", alertBody["title"])
	ev.Payload.Topic = payload.Topic

	return nil
}

func (event *ApnsEvent) report(kafkaProducer kafka.Producer, kafkaReportingTopic string) error {
	if kafkaProducer == nil || kafkaReportingTopic == "" {
		return errApnsKafkaReportingConfiguration
	}
	uuid, err := go_uuid.New()
	if err != nil {
		logger.WithError(err).Error("Could not get new UUID")
		return err
	}
	responseTime := time.Now().UTC().Format(time.RFC3339)
	event.ID = uuid
	event.Time = responseTime

	bytesReportEvent, err := json.Marshal(event)
	if err != nil {
		logger.WithError(err).Error("Error while marshaling Kafka reporting event to JSON format")
		return err
	}
	logger.WithField("event", *event).Debug("Reporting sent APNS event  to Kafka topic")
	kafkaProducer.Report(kafkaReportingTopic, bytesReportEvent, uuid)
	return nil
}

/*
	This is a copy of the apns2 format in order to decapsulate the data from the gobbler message.
*/
type Payload struct {
	Deeplink string `json:"deeplink"`
	Topic    string `json:"topic"`
	Aps      Aps    `json:"aps"`
}

type Aps struct {
	Alert            interface{} `json:"alert,omitempty"`
	Badge            interface{} `json:"badge,omitempty"`
	Category         string      `json:"category,omitempty"`
	ContentAvailable int         `json:"content-available,omitempty"`
	MutableContent   int         `json:"mutable-content,omitempty"`
	Sound            string      `json:"sound,omitempty"`
	ThreadID         string      `json:"thread-id,omitempty"`
	URLArgs          []string    `json:"url-args,omitempty"`
}
