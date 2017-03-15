package sms

import (
	"encoding/json"
	"fmt"
	"github.com/cosminrentea/gobbler/testutil"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

func Test_NexmoHTTPError(t *testing.T) {
	defer testutil.EnableDebugForMethod()()
	a := assert.New(t)

	port := createRandomPort(7000, 9000)
	URL = "http://127.0.0.1" + port

	expectedRequestNo := 3
	go dummyNexmoEndpointWithHandlerFunc(t, &expectedRequestNo, port, noResponseFromNexmoHandler)
	kvStore, f := createKVStore(t, "/guble_sms_nexmo_error")
	defer os.Remove(f)

	gw := createGateway(t, kvStore)

	msg := encodeProtocolMessage(t, 2)

	err := gw.route.Deliver(&msg, false)
	a.NoError(err)
	time.Sleep(5 * timeInterval)
	a.Equal(0, expectedRequestNo, "Three retries should be made by sender.")
	a.Equal(msg.ID, gw.LastIDSent, "Retry failed.Last id  sent should be msgId")

	stopGateway(t, gw)
}

func Test_NexmoInvalidSenderError(t *testing.T) {
	defer testutil.EnableDebugForMethod()()
	a := assert.New(t)

	port := createRandomPort(7000, 9000)
	URL = "http://127.0.0.1" + port

	expectedRequestNo := 1
	go dummyNexmoEndpointWithHandlerFunc(t, &expectedRequestNo, port, invalidSenderNexmoHandler)
	kvStore, f := createKVStore(t, "/guble_sms_nexmo_invalid_sender_error")
	defer os.Remove(f)

	gw := createGateway(t, kvStore)

	msg := encodeProtocolMessage(t, 2)
	err := gw.route.Deliver(&msg, false)
	a.NoError(err)
	time.Sleep(2 * timeInterval)
	a.Equal(0, expectedRequestNo, "Only one try should be made by sender.")
	a.Equal(msg.ID, gw.LastIDSent, "No Retry needed.Last id  sent should be msgId")

	stopGateway(t, gw)
}

func Test_NexmoMultipleErrorsFollowedBySuccess(t *testing.T) {

}

func Test_NexmoResponseCodeError(t *testing.T) {
	defer testutil.EnableDebugForMethod()()
	a := assert.New(t)

	port := createRandomPort(7000, 9000)
	URL = "http://127.0.0.1" + port

	expectedRequestNo := 3
	go dummyNexmoEndpointWithHandlerFunc(t, &expectedRequestNo, port, responseInternalErrorNexmoHandler)
	kvStore, f := createKVStore(t, "/guble_sms_nexmo_responde_code_error")
	defer os.Remove(f)

	gw := createGateway(t, kvStore)

	msg := encodeProtocolMessage(t, 2)
	err := gw.route.Deliver(&msg, false)
	a.NoError(err)
	time.Sleep(5 * timeInterval)
	a.Equal(0, expectedRequestNo, "Only one try should be made by sender.")
	a.Equal(msg.ID, gw.LastIDSent, "No Retry needed.Last id  sent should be msgId")

	stopGateway(t, gw)
}

func noResponseFromNexmoHandler(t *testing.T, noOfReq *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := assert.New(t)
		sentSms := decodeSMSMessage(t, r)
		a.Equal("body", sentSms.Text)
		*noOfReq--
	}
}

func invalidSenderNexmoHandler(t *testing.T, noOfReq *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := assert.New(t)
		sentSms := decodeSMSMessage(t, r)
		a.Equal("body", sentSms.Text)
		*noOfReq--

		nexmoResponse := composeNexmoMessageResponse(sentSms, ResponseInvalidSenderAddress, 1)
		writeNexmoResponse(nexmoResponse, t, w)

	}
}

func responseInternalErrorNexmoHandler(t *testing.T, noOfReq *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := assert.New(t)
		sentSms := decodeSMSMessage(t, r)
		a.Equal("body", sentSms.Text)
		*noOfReq--

		nexmoResponse := composeNexmoMessageResponse(sentSms, ResponseInternalError, 1)

		writeNexmoResponse(nexmoResponse, t, w)

	}
}

func dummyNexmoEndpointWithHandlerFunc(t *testing.T, expectedRequestNo *int, port string, handler func(t *testing.T, no *int) http.HandlerFunc) {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", handler(t, expectedRequestNo))
	fmt.Println(port)
	http.ListenAndServe(port, serveMux)
}

func decodeSMSMessage(t *testing.T, r *http.Request) NexmoSms {
	a := assert.New(t)
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	var sentSms NexmoSms
	err := json.Unmarshal(body, &sentSms)
	if err != nil {
		a.FailNow("Could not decode sender sms.")
	}
	return sentSms
}
func writeNexmoResponse(nexmoResponse *NexmoMessageResponse, t *testing.T, w http.ResponseWriter) {
	a := assert.New(t)
	jData, err := json.Marshal(nexmoResponse)
	if err != nil {
		a.FailNow("Nexmo Response encoding failed.")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}
