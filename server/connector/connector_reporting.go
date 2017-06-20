package connector

import (
	"encoding/json"
	"errors"
	"github.com/cosminrentea/go-uuid"
	"github.com/cosminrentea/gobbler/server/kafka"
	"time"
)

const (
	deviceTokenKey = "device_token"
	userIDKEy      = "user_id"
)

type SubscribeUnsubscribePayload struct {
	Service   string `json:"service"`
	Topic     string `json:"topic"`
	DeviceID  string `json:"device_id"`
	UserID    string `json:"user_id"`
	Action    string `json:"action"`
	ErrorText string `json:"error_text"`
}

type SubscribeUnsubscribeEvent struct {
	Id      string                      `json:"id"`
	Time    string                      `json:"time"`
	Type    string                      `json:"type"`
	Payload SubscribeUnsubscribePayload `json:"payload"`
}

var (
	errKafkaReportingConfiguration = errors.New("Kafka Reporting for Subscribe/Unsubscribe is not correctly configured")
	errInvalidParams               = errors.New("Could not extract params")
)

func (event *SubscribeUnsubscribeEvent) report(kafkaProducer kafka.Producer, kafkaReportingTopic string) error {
	if kafkaProducer == nil || kafkaReportingTopic == "" {
		return errKafkaReportingConfiguration
	}
	uuid, err := go_uuid.New()
	if err != nil {
		logger.WithError(err).Error("Could not get new UUID")
		return err
	}
	responseTime := time.Now().UTC().Format(time.RFC3339)
	event.Id = uuid
	event.Time = responseTime

	bytesReportEvent, err := json.Marshal(event)
	if err != nil {
		logger.WithError(err).Error("Error while marshaling Kafka reporting event to JSON format")
		return err
	}
	logger.WithField("event", *event).Debug("Reporting sent subscribe unsubscribe to Kafka topic")
	kafkaProducer.Report(kafkaReportingTopic, bytesReportEvent, uuid)
	return nil
}

func (event *SubscribeUnsubscribeEvent) fillParams(params map[string]string) error {
	deviceID, ok := params[deviceTokenKey]
	if !ok {
		return errInvalidParams
	}
	event.Payload.DeviceID = deviceID

	userID, ok := params[userIDKEy]
	if !ok {
		return errInvalidParams
	}
	event.Payload.UserID = userID
	return nil
}
