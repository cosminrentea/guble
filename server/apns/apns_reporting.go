package apns

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/cosminrentea/go-uuid"
	"github.com/cosminrentea/gobbler/server/kafka"
)

type ApnsEventPayload struct {
	Topic             string `json:"topic"`
	Status            string `json:"status"`
	ErrorText         string `json:"error_text"`
	ApnsID            string `json:"apns_id"`
	CorrelationID     string `json:"correlation_id"`
	UserID            string `json:"user_id"`
	DeviceID          string `json:"device_id"`
	NotificationBody  string `json:"notification_body"`
	NotificationTitle string `json:"notification_title"`
	DeepLink          string `json:"deep_link"`
}

type ApnsEvent struct {
	Id      string           `json:"id"`
	Time    string           `json:"time"`
	Type    string           `json:"type"`
	Payload ApnsEventPayload `json:"payload"`
}

var (
	errApnsKafkaReportingConfiguration = errors.New("Kafka Reporting for APNS is not correctly configured")
)

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
	event.Id = uuid
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
