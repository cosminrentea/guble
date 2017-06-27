package kafka

type PushEventPayload struct {
	Topic             string `json:"topic"`
	Status            string `json:"status"`
	ErrorText         string `json:"error_text"`
	UserID            string `json:"user_id"`
	DeviceID          string `json:"device_id"`
	NotificationBody  string `json:"notification_body"`
	NotificationTitle string `json:"notification_title"`
	DeepLink          string `json:"deep_link"`
}
