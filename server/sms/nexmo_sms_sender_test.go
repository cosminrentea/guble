package sms

import (
	"net/http"
	"testing"
	"time"

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
	RequestTimeout = time.Second
	a := assert.New(t)
	sender, err := NewNexmoSender(KEY, SECRET)
	a.NoError(err)

	msg := encodeProtocolMessage(t, 0)

	err = sender.Send(&msg)
	time.Sleep(3 * timeInterval)
	a.Error(err)
	a.Equal(ErrRetryFailed, err)
}

func TestNexmoSender_SendExpiredMessage(t *testing.T) {
	a := assert.New(t)

	port := createRandomPort(7000, 8000)
	URL = "http://127.0.0.1" + port

	sender := createNexmoSender(t)
	countCh := make(chan bool)
	doneCh := make(chan struct{})
	// no request should be made in case the sms is expired
	go dummyNexmoEndpointWithHandlerFunc(t, countCh, port, func(t *testing.T, countCh chan bool) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			countCh <- true
			response := composeNexmoMessageResponse(decodeSMSMessage(t, r), ResponseSuccess, 1)
			writeNexmoResponse(response, t, w)
		}
	})
	go func() {
		select {
		case <-countCh:
			a.FailNow("Nexmo call not expected.")
		case <-doneCh:
			return
		}
	}()

	msg := encodeProtocolMessage(t, 0)
	expires := time.Now().Add(-1 * time.Hour)
	msg.Expires = &expires

	err := sender.Send(&msg)
	time.Sleep(3 * timeInterval)
	a.NoError(err)
	close(doneCh)
}
