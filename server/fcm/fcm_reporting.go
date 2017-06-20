package fcm

import (
	"encoding/json"
	"errors"
	"time"

	"fmt"
	"github.com/Bogh/gcm"
	"github.com/cosminrentea/go-uuid"
	"github.com/cosminrentea/gobbler/server/connector"
	"github.com/cosminrentea/gobbler/server/kafka"
)

type FcmEventPayload struct {
	Topic             string `json:"topic"`
	Status            string `json:"status"`
	ErrorText         string `json:"error_text"`
	CanonicalID       string `json:"canonical_id"`
	CorrelationID     string `json:"correlation_id"`
	UserID            string `json:"user_id"`
	DeviceID          string `json:"device_id"`
	NotificationBody  string `json:"notification_body"`
	NotificationTitle string `json:"notification_title"`
	DeepLink          string `json:"deep_link"`
}

type FcmEvent struct {
	Id      string          `json:"id"`
	Time    string          `json:"time"`
	Type    string          `json:"type"`
	Payload FcmEventPayload `json:"payload"`
}

var (
	errFcmKafkaReportingConfiguration = errors.New("Kafka Reporting for FCM is not correctly configured")
	errFcmMessageDecodingFailed       = errors.New("Decoding of fcm payload field failed")
)

func (ev *FcmEvent) fillApnsEvent(request connector.Request) error {

	ev.Payload.CorrelationID = request.Message().CorrelationID()

	deviceID := request.Subscriber().Route().Get(deviceTokenKey)
	ev.Payload.DeviceID = deviceID

	userID := request.Subscriber().Route().Get(userIDKEy)
	ev.Payload.UserID = userID

	var msg gcm.Message
	err := json.Unmarshal(request.Message().Body, &msg)
	if err != nil {
		logger.WithError(err).Error("Error reading msg notification built.")
		return errFcmMessageDecodingFailed
	}

	ev.Payload.DeepLink = fmt.Sprintf("%s", msg.Data["deep_link"])
	ev.Payload.NotificationBody = fmt.Sprintf("%s", msg.Data["notification_body"])
	ev.Payload.NotificationTitle = fmt.Sprintf("%s", msg.Data["notification_title"])
	ev.Payload.Topic = fmt.Sprintf("%s", msg.Data["type"])

	return nil
}

func (event *FcmEvent) report(kafkaProducer kafka.Producer, kafkaReportingTopic string) error {
	if kafkaProducer == nil || kafkaReportingTopic == "" {
		return errFcmKafkaReportingConfiguration
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
	logger.WithField("event", *event).Debug("Reporting sent FCM event  to Kafka topic")
	kafkaProducer.Report(kafkaReportingTopic, bytesReportEvent, uuid)
	return nil
}
