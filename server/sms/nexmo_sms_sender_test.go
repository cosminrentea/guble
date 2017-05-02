package sms

import (
	"testing"
	"time"

	"github.com/cosminrentea/gobbler/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_HttpClientRecreation(t *testing.T) {
	a := assert.New(t)

	port := createRandomPort(7000, 8000)
	URL = "http://127.0.0.1" + port

	sender := createNexmoSender(t)
	msg := encodeProtocolMessage(t, 2)

	err := sender.Send(&msg)
	time.Sleep(3 * timeInterval)

	a.Equal(ErrRetryFailed, err)
}

func TestNexmoSender_SendWithError(t *testing.T) {
	defer testutil.EnableDebugForMethod()
	RequestTimeout = time.Second
	a := assert.New(t)
	sender, err := NewNexmoSender(KEY, SECRET,nil,"")
	a.NoError(err)

	msg := encodeProtocolMessage(t, 0)

	err = sender.Send(&msg)
	time.Sleep(3 * timeInterval)
	a.Error(err)
	a.Equal(ErrRetryFailed, err)
}

func TestNexmoSender_SendReport(t *testing.T) {
	ctrl, finish := testutil.NewMockCtrl(t)
	defer finish()
	a := assert.New(t)
	defer testutil.EnableDebugForMethod()()

	port := createRandomPort(7000, 9000)
	URL = "http://127.0.0.1" + port

	countCh := make(chan bool, 0)
	go dummyNexmoEndpointWithHandlerFunc(t, countCh, port, succesSenderNexmoHandler)
	producer := NewMockProducer(ctrl)
	sender, err := NewNexmoSender(KEY, SECRET, producer, "sms_reporting")
	a.NoError(err)

	msg := encodeProtocolMessage(t, 0)
	producer.EXPECT().Report(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(topic string, bytes []byte, key string) {
		a.Equal("sms_reporting", topic)
	})

	err = sender.Send(&msg)
	a.NoError(err)

	readServedResponses(t,1,countCh)
}
