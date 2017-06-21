package apns

import (
	"errors"
	"testing"
	"time"

	"encoding/json"

	"github.com/cosminrentea/gobbler/protocol"
	"github.com/cosminrentea/gobbler/server/connector"
	"github.com/cosminrentea/gobbler/server/kafka"
	"github.com/cosminrentea/gobbler/server/router"
	"github.com/cosminrentea/gobbler/testutil"
	"github.com/golang/mock/gomock"
	"github.com/sideshow/apns2"
	_ "github.com/sideshow/apns2/payload"
	"github.com/stretchr/testify/assert"
)

var ErrSendRandomError = errors.New("A Sender error")

func TestNew_WithoutKVStore(t *testing.T) {
	_, finish := testutil.NewMockCtrl(t)
	defer finish()
	a := assert.New(t)

	//given
	mRouter := NewMockRouter(testutil.MockCtrl)
	errKVS := errors.New("No KVS was set-up in Router")
	mRouter.EXPECT().KVStore().Return(nil, errKVS).AnyTimes()
	mSender := NewMockSender(testutil.MockCtrl)
	prefix := "/apns/"
	workers := 1
	cfg := Config{
		Prefix:  &prefix,
		Workers: &workers,
	}

	//when
	c, err := New(mRouter, mSender, cfg, nil, "sub_kafka_reporting", "apns_Reporting")

	//then
	a.Error(err)
	a.Nil(c)
}

func TestConn_HandleResponseOnSendError(t *testing.T) {
	_, finish := testutil.NewMockCtrl(t)
	defer finish()
	defer testutil.EnableDebugForMethod()()
	a := assert.New(t)

	//given
	c, _ := newAPNSConnector(t,nil)
	mRequest := NewMockRequest(testutil.MockCtrl)
	message := &protocol.Message{
		HeaderJSON: `{"Correlation-Id": "7sdks723ksgqn"}`,
		ID:         42,
	}
	route := testRoute()

	mRequest.EXPECT().Message().Return(message).AnyTimes()
	mSubscriber := NewMockSubscriber(testutil.MockCtrl)
	mRequest.EXPECT().Subscriber().Return(mSubscriber).AnyTimes()
	mSubscriber.EXPECT().Route().Return(route).AnyTimes()

	time.Sleep(100 * time.Millisecond)
	//when
	err := c.HandleResponse(mRequest, nil, nil, ErrSendRandomError)

	//then
	a.Equal(ErrSendRandomError, err)
}

func TestConn_HandleResponse(t *testing.T) {
	_, finish := testutil.NewMockCtrl(t)
	defer finish()
	defer testutil.EnableDebugForMethod()()
	a := assert.New(t)

	//given
	c, mKVS := newAPNSConnector(t, nil)

	route := testRoute()

	mSubscriber := NewMockSubscriber(testutil.MockCtrl)
	mSubscriber.EXPECT().SetLastID(gomock.Any())
	mSubscriber.EXPECT().Key().Return("key").AnyTimes()
	mSubscriber.EXPECT().Encode().Return([]byte("{}"), nil).AnyTimes()
	mSubscriber.EXPECT().Route().Return(route).AnyTimes()
	mKVS.EXPECT().Put(schema, "key", []byte("{}")).AnyTimes()

	c.Manager().Add(mSubscriber)
	message := &protocol.Message{
		ID:         42,
		HeaderJSON: `{"Content-Type": "text/plain", "Correlation-Id": "7sdks723ksgqn"}`,
		Body:       []byte("{}"),
	}

	mRequest := NewMockRequest(testutil.MockCtrl)
	mRequest.EXPECT().Message().Return(message).AnyTimes()
	mRequest.EXPECT().Subscriber().Return(mSubscriber).AnyTimes()
	mSubscriber.EXPECT().Route().Return(route).AnyTimes()
	response := &apns2.Response{
		ApnsID:     "id-life",
		StatusCode: 200,
	}

	//when
	err := c.HandleResponse(mRequest, response, nil, nil)

	//then
	a.NoError(err)
}

func TestNew_HandleResponseHandleSubscriber(t *testing.T) {
	_, finish := testutil.NewMockCtrl(t)
	defer finish()
	defer testutil.EnableDebugForMethod()()
	a := assert.New(t)

	//given
	c, mKVS := newAPNSConnector(t,nil)

	removeForReasons := []string{
		apns2.ReasonMissingDeviceToken,
		apns2.ReasonBadDeviceToken,
		apns2.ReasonDeviceTokenNotForTopic,
		apns2.ReasonUnregistered,
	}
	route := testRoute()
	for _, reason := range removeForReasons {
		message := &protocol.Message{
			ID:         42,
			HeaderJSON: `{"Correlation-Id": "7sdks723ksgqn"}`,
		}
		mSubscriber := NewMockSubscriber(testutil.MockCtrl)
		mSubscriber.EXPECT().SetLastID(gomock.Any())
		mSubscriber.EXPECT().Cancel()
		mSubscriber.EXPECT().Key().Return("key").AnyTimes()
		mSubscriber.EXPECT().Encode().Return([]byte("{}"), nil).AnyTimes()
		mSubscriber.EXPECT().Route().Return(route).AnyTimes()
		mKVS.EXPECT().Put(schema, "key", []byte("{}")).Times(2)
		mKVS.EXPECT().Delete(schema, "key")

		c.Manager().Add(mSubscriber)

		mRequest := NewMockRequest(testutil.MockCtrl)
		mRequest.EXPECT().Message().Return(message).AnyTimes()
		mRequest.EXPECT().Subscriber().Return(mSubscriber).AnyTimes()

		response := &apns2.Response{
			ApnsID:     "id-life",
			StatusCode: 400,
			Reason:     reason,
		}

		//when
		err := c.HandleResponse(mRequest, response, nil, nil)

		//then
		a.NoError(err)
	}
}

func TestNew_HandleResponseDoNotHandleSubscriber(t *testing.T) {
	_, finish := testutil.NewMockCtrl(t)
	defer finish()
	a := assert.New(t)

	//given
	c, mKVS := newAPNSConnector(t,nil)
	route := testRoute()

	noActionForReasons := []string{
		apns2.ReasonPayloadEmpty,
		apns2.ReasonPayloadTooLarge,
		apns2.ReasonBadTopic,
		apns2.ReasonTopicDisallowed,
		apns2.ReasonBadMessageID,
		apns2.ReasonBadExpirationDate,
		apns2.ReasonBadPriority,
		apns2.ReasonDuplicateHeaders,
		apns2.ReasonBadCertificateEnvironment,
		apns2.ReasonBadCertificate,
		apns2.ReasonForbidden,
		apns2.ReasonBadPath,
		apns2.ReasonMethodNotAllowed,
		apns2.ReasonTooManyRequests,
		apns2.ReasonIdleTimeout,
		apns2.ReasonShutdown,
		apns2.ReasonInternalServerError,
		apns2.ReasonServiceUnavailable,
		apns2.ReasonMissingTopic,
	}

	for _, reason := range noActionForReasons {
		message := &protocol.Message{
			ID: 42,
		}

		mSubscriber := NewMockSubscriber(testutil.MockCtrl)
		mSubscriber.EXPECT().SetLastID(gomock.Any())
		mSubscriber.EXPECT().Key().Return("key").AnyTimes()
		mSubscriber.EXPECT().Encode().Return([]byte("{}"), nil).AnyTimes()
		mSubscriber.EXPECT().Cancel()
		mSubscriber.EXPECT().Route().Return(route).AnyTimes()
		mKVS.EXPECT().Put(schema, "key", []byte("{}")).Times(2)
		mKVS.EXPECT().Delete(schema, "key")

		c.Manager().Add(mSubscriber)

		mRequest := NewMockRequest(testutil.MockCtrl)
		mRequest.EXPECT().Message().Return(message).AnyTimes()
		mRequest.EXPECT().Subscriber().Return(mSubscriber).AnyTimes()

		response := &apns2.Response{
			ApnsID:     "id-apns",
			StatusCode: 400,
			Reason:     reason,
		}

		//when
		err := c.HandleResponse(mRequest, response, nil, nil)

		//then
		a.NoError(err)

		c.Manager().Remove(mSubscriber)
	}
}

func newAPNSConnector(t *testing.T, producer kafka.Producer) (c connector.ResponsiveConnector, mKVS *MockKVStore) {
	mKVS = NewMockKVStore(testutil.MockCtrl)
	mRouter := NewMockRouter(testutil.MockCtrl)
	mRouter.EXPECT().KVStore().Return(mKVS, nil).AnyTimes()
	mSender := NewMockSender(testutil.MockCtrl)

	prefix := "/apns/"
	workers := 1
	intervalMetrics := false
	password := "test"
	bytes := []byte("test")
	cfg := Config{
		Prefix:              &prefix,
		Workers:             &workers,
		IntervalMetrics:     &intervalMetrics,
		CertificatePassword: &password,
		CertificateBytes:    &bytes,
	}

	c, err := New(mRouter, mSender, cfg, producer, "sub_reporting", "apns_Reporting")
	assert.NoError(t, err)
	assert.NotNil(t, c)

	return
}

func testRoute() *router.Route {
	options := router.RouteConfig{
		RouteParams: router.RouteParams{
			deviceIDKey: "device_id",
			userIDKey:   "user_id",
		},
	}
	return router.NewRoute(options)
}

func TestConn_HandleResponseReporting(t *testing.T) {
	ctrl, finish := testutil.NewMockCtrl(t)
	//defer testutil.EnableDebugForMethod()()
	defer finish()
	a := assert.New(t)

	mockProducer := NewMockProducer(ctrl)

	//given
	c, mKVS := newAPNSConnector(t, mockProducer)

	route := testRoute()

	mSubscriber := NewMockSubscriber(testutil.MockCtrl)
	mSubscriber.EXPECT().SetLastID(gomock.Any())
	mSubscriber.EXPECT().Key().Return("key").AnyTimes()
	mSubscriber.EXPECT().Encode().Return([]byte("{}"), nil).AnyTimes()
	mSubscriber.EXPECT().Route().Return(route).AnyTimes()
	mKVS.EXPECT().Put(schema, "key", []byte("{}")).Times(2)

	c.Manager().Add(mSubscriber)
	message := &protocol.Message{
		UserID:     "user_id",
		ID:         42,
		HeaderJSON: `{"Content-Type": "text/plain", "Correlation-Id": "7sdks723ksgqn"}`,
		Body: []byte(`{
		"aps":{
		"alert":{"body":"Die größte Sonderangebot!","title":"Valid Title"},
		"badge":0,
		"content-available":1
		},
		"topic":"marketing_notifications",
		"deeplink":"rewe://angebote"
		}`),
	}
	mRequest := NewMockRequest(testutil.MockCtrl)
	mRequest.EXPECT().Message().Return(message).AnyTimes()
	mRequest.EXPECT().Subscriber().Return(mSubscriber).AnyTimes()

	response := &apns2.Response{
		ApnsID:     "apns_id",
		StatusCode: 200,
	}

	mockProducer.EXPECT().Report(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(topic string, bytes []byte, key string) {
		a.Equal("apns_Reporting", topic)

		var event ApnsEvent
		err := json.Unmarshal(bytes, &event)
		a.NoError(err)
		a.Equal("pn_reporting_apns", event.Type)
		a.Equal("Success", event.Payload.Status)
		a.Equal("apns_id", event.Payload.ApnsID)
		a.Equal("7sdks723ksgqn", event.Payload.CorrelationID)
		a.Equal("device_id", event.Payload.DeviceID)
		a.Equal("user_id", event.Payload.UserID)
		a.Equal("Valid Title", event.Payload.NotificationTitle)
		a.Equal("Die größte Sonderangebot!", event.Payload.NotificationBody)
		a.Equal("rewe://angebote", event.Payload.DeepLink)
		a.Equal("marketing_notifications", event.Payload.Topic)
		a.Equal("", event.Payload.ErrorText)
	})

	//when
	err := c.HandleResponse(mRequest, response, nil, nil)

	//then
	a.NoError(err)
}
