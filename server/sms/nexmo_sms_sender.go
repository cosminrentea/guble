package sms

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cosminrentea/go-uuid"
	"github.com/cosminrentea/gobbler/protocol"
	"github.com/cosminrentea/gobbler/server/kafka"
	"github.com/jpillora/backoff"
)

type ResponseCode int

const (
	ResponseSuccess ResponseCode = iota
	ResponseThrottled
	ResponseMissingParams
	ResponseInvalidParams
	ResponseInvalidCredentials
	ResponseInternalError
	ResponseInvalidMessage
	ResponseNumberBarred
	ResponsePartnerAcctBarred
	ResponsePartnerQuotaExceeded
	ResponseUnused
	ResponseRESTNotEnabled
	ResponseMessageTooLong
	ResponseCommunicationFailed
	ResponseInvalidSignature
	ResponseInvalidSenderAddress
	ResponseInvalidTTL
	ResponseFacilityNotAllowed
	ResponseInvalidMessageClass
)

var (
	URL                = "https://rest.nexmo.com/sms/json?"
	MaxIdleConnections = 100
	RequestTimeout     = 500 * time.Millisecond

	ErrHTTPClientError           = errors.New("Http client sending to Nexmo Failed.No sms was sent")
	ErrNexmoResponseStatusNotOk  = errors.New("Nexmo response status not ResponseSuccess")
	ErrSMSResponseDecodingFailed = errors.New("Nexmo response decoding failed")
	ErrInvalidSender             = errors.New("Sms destination phoneNumber is invalid")
	ErrMultipleSmsSent           = errors.New("Multiple  or no sms we're sent.SMS message may be too long")
	ErrRetryFailed               = errors.New("Failed retrying to send message")
	ErrEncodeFailed              = errors.New("Encoding of message to be sent to Nexmo  failed")

	errKafkaReportingConfiguration = errors.New("Kafka Reporting for Nexmo is not correctly configured")
)

var nexmoResponseCodeMap = map[ResponseCode]string{
	ResponseSuccess:              "Success",
	ResponseThrottled:            "Throttled",
	ResponseMissingParams:        "Missing params",
	ResponseInvalidParams:        "Invalid params",
	ResponseInvalidCredentials:   "Invalid credentials",
	ResponseInternalError:        "Internal error",
	ResponseInvalidMessage:       "Invalid message",
	ResponseNumberBarred:         "Number barred",
	ResponsePartnerAcctBarred:    "Partner account barred",
	ResponsePartnerQuotaExceeded: "Partner quota exceeded",
	ResponseRESTNotEnabled:       "Account not enabled for REST",
	ResponseMessageTooLong:       "Message too long",
	ResponseCommunicationFailed:  "Communication failed",
	ResponseInvalidSignature:     "Invalid signature",
	ResponseInvalidSenderAddress: "Invalid sender address",
	ResponseInvalidTTL:           "Invalid TTL",
	ResponseFacilityNotAllowed:   "Facility not allowed",
	ResponseInvalidMessageClass:  "Invalid message class",
}

func (c ResponseCode) String() string {
	return nexmoResponseCodeMap[c]
}

// NexmoMessageReport is the "status report" for a single SMS sent via the Nexmo API
type NexmoMessageReport struct {
	Status           ResponseCode `json:"status,string"`
	MessageID        string       `json:"message-id"`
	To               string       `json:"to"`
	ClientReference  string       `json:"client-ref"`
	RemainingBalance string       `json:"remaining-balance"`
	MessagePrice     string       `json:"message-price"`
	Network          string       `json:"network"`
	ErrorText        string       `json:"error-text"`
}

type ReportPayload struct {
	OrderId         int                `json:"order_id"`
	MessageId       string             `json:"message_id"`
	SmsText         string             `json:"sms_text"`
	SmsRequestTime  string             `json:"sms_request_time"`
	SmsResponseTime string             `json:"sms_response_time"`
	SmsResponse     NexmoMessageReport `json:"sms_response"`
}

type ReportEvent struct {
	Id      string        `json:"id"`
	Time    string        `json:"time"`
	Type    string        `json:"type"`
	Payload ReportPayload `json:"payload"`
}

type NexmoMessageResponse struct {
	MessageCount int                  `json:"message-count,string"`
	Messages     []NexmoMessageReport `json:"messages"`
}

func (nm NexmoMessageResponse) Check() error {
	if nm.MessageCount != 1 {
		logger.WithField("message_count", nm.MessageCount).Error("Nexmo message count error.")
		return ErrMultipleSmsSent
	}
	if nm.Messages[0].Status != ResponseSuccess {
		logger.WithField("status", nm.Messages[0].Status).WithField("error", nm.Messages[0].ErrorText).
			Error("Error received from Nexmo")

		if nm.Messages[0].Status == ResponseInvalidSenderAddress {
			logger.Info("Invalid Sender detected.No retries will be made.")
			return ErrInvalidSender
		}
		return ErrNexmoResponseStatusNotOk
	}
	return nil
}

type NexmoSender struct {
	logger *log.Entry

	ApiKey    string
	ApiSecret string

	kafkaProducer       kafka.Producer
	kafkaReportingTopic string

	httpClient *http.Client
}

func NewNexmoSender(apiKey, apiSecret string, kafkaProducer kafka.Producer, kafkaReportingTopic string) (*NexmoSender, error) {
	ns := &NexmoSender{
		logger:              logger.WithField("name", "nexmoSender"),
		ApiKey:              apiKey,
		ApiSecret:           apiSecret,
		kafkaProducer:       kafkaProducer,
		kafkaReportingTopic: kafkaReportingTopic,
	}
	ns.createHttpClient()
	return ns, nil
}

func (ns *NexmoSender) Send(msg *protocol.Message) error {
	nexmoSMS := new(NexmoSms)
	err := json.Unmarshal(msg.Body, nexmoSMS)
	if err != nil {
		logger.WithField("msg", msg).WithField("error", err.Error()).Error("Could not decode message body to send to nexmo.No retries will be made for this message.")
		return ErrRetryFailed
	}

	withRetry := &retryable{
		maxTries: 3,
		Backoff: backoff.Backoff{
			Min:    50 * time.Millisecond,
			Max:    250 * time.Millisecond,
			Factor: 2,
			Jitter: true,
		},
	}
	err = withRetry.executeAndCheck(
		func() (*NexmoMessageResponse, error) {
			return ns.sendSms(nexmoSMS)
		},
		ns.kafkaProducer,
		ns.kafkaReportingTopic,
	)
	if err == ErrRetryFailed {
		logger.WithField("msg", msg).Info("Retry failed or not necessary.Moving on")
	}

	return err
}

type retryable struct {
	backoff.Backoff
	maxTries int
}

func (r *retryable) executeAndCheck(op func() (*NexmoMessageResponse, error), kafkaProducer kafka.Producer, kafkaReportingTopic string) error {
	tryCounter := 0
	for {
		tryCounter++
		nexmoSMSResponse, err := op()
		if err == nil {
			logger.WithField("response", nexmoSMSResponse).WithField("try", tryCounter).Info("Decoded nexmo response")
			err = nexmoSMSResponse.Check()
			if err == nil {
				err = report(nexmoSMSResponse.Messages[0], kafkaProducer, kafkaReportingTopic)
				if err != nil {
					logger.WithError(err).Error("Could not report sent nexmo sms to Kafka topic")
				}
				return nil
			}
			if err == ErrInvalidSender {
				return ErrRetryFailed
			}
		}
		if tryCounter >= r.maxTries {
			return ErrRetryFailed
		}
		d := r.Duration()
		logger.WithField("error", err.Error()).WithField("duration", d).Info("Retry in")
		time.Sleep(d)
	}
}

func report(nexmoMessageReport NexmoMessageReport, kafkaProducer kafka.Producer, kafkaReportingTopic string) error {
	if kafkaProducer == nil || kafkaReportingTopic == "" {
		return errKafkaReportingConfiguration
	}
	uuid, err := go_uuid.New()
	if err != nil {
		return err
	}
	bytesReportEvent, err := json.Marshal(ReportEvent{
		Id:   uuid,
		Time: time.Now().UTC().Format(time.RFC3339),
		Type: "tour_arrival_estimate_nexmo",
		Payload: ReportPayload{
			SmsResponse: nexmoMessageReport,
		},
	})
	if err != nil {
		return err
	}
	logger.Debug("Reporting sent nexmo sms to Kafka topic")
	kafkaProducer.Report(kafkaReportingTopic, bytesReportEvent)
	return nil
}

func (ns *NexmoSender) sendSms(sms *NexmoSms) (*NexmoMessageResponse, error) {
	logger.WithField("sms_details", sms).WithField("order_id", sms.ClientRef).Info("sendSms")

	smsEncoded, err := sms.EncodeNexmoSms(ns.ApiKey, ns.ApiSecret)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Error encoding sms")
		return nil, ErrEncodeFailed
	}

	req, err := http.NewRequest(http.MethodPost, URL, bytes.NewBuffer(smsEncoded))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(smsEncoded)))

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Error doing the request to nexmo endpoint")
		ns.createHttpClient()
		mTotalSendErrors.Add(1)
		pNexmoSendErrors.Inc()
		return nil, ErrHTTPClientError
	}
	defer resp.Body.Close()

	var messageResponse *NexmoMessageResponse
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Error reading the nexmo body response")
		mTotalResponseInternalErrors.Add(1)
		pNexmoResponseInternalErrors.Inc()
		return nil, ErrSMSResponseDecodingFailed
	}

	err = json.Unmarshal(respBody, &messageResponse)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Error decoding the response from nexmo endpoint")
		mTotalResponseInternalErrors.Add(1)
		pNexmoResponseInternalErrors.Inc()
		return nil, ErrSMSResponseDecodingFailed
	}
	logger.WithField("messageResponse", messageResponse).WithField("csOrderId", sms.ClientRef).Info("Actual nexmo response")

	return messageResponse, nil
}

func (ns *NexmoSender) createHttpClient() {
	logger.Info("Recreating HTTP client for nexmo sender")
	ns.httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: MaxIdleConnections,
		},
		Timeout: RequestTimeout,
	}
}
