package sms

import (
	"testing"
	"time"

	"github.com/cosminrentea/gobbler/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_HttpClientRecreation(t *testing.T) {
	defer testutil.EnableDebugForMethod()()
	a := assert.New(t)

	port := createRandomPort(7000, 8000)
	URL = "http://127.0.0.1" + port
	expectedRequestNo := 3

	go dummyNexmoEndpointWithHandlerFunc(t, &expectedRequestNo, port, noResponseFromNexmoHandler)

	sender := createNexmoSender(t)
	msg := encodeProtocolMessage(t, 2)

	err := sender.Send(&msg)
	a.Equal(ErrRetryFailed, err)

	a.Equal(0, expectedRequestNo, "Three retries should be made by sender.")
	time.Sleep(timeInterval)
}

func TestNexmoSender_SendWithError(t *testing.T) {
	defer testutil.EnableDebugForMethod()
	RequestTimeout = time.Second
	a := assert.New(t)
	sender, err := NewNexmoSender(KEY, SECRET)
	a.NoError(err)

	msg := encodeProtocolMessage(t, 0)

	err = sender.Send(&msg)
	a.Error(err)
	a.Equal(ErrRetryFailed, err)
}
