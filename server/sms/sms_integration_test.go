package sms

import (
	"encoding/json"
	"github.com/cosminrentea/gobbler/protocol"
	"github.com/cosminrentea/gobbler/testutil"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func Test_HttpClientRecreation(t *testing.T) {
	defer testutil.EnableDebugForMethod()()
	a := assert.New(t)

	port := ":9001"
	URL = "http://127.0.0.1" + port
	expectedRequestNo :=3

	go dummyNexmoEndpointWithHandlerFunc(t,&expectedRequestNo, port, noResponseFromNexmoHandler)

	sender := createNexmoSender(t)
	msg := encodeProtocolMessage(t)

	err := sender.Send(&msg)
	a.NoError(err)

	a.Equal(0,expectedRequestNo,"Three retries should be made by sender.")

}

func encodeProtocolMessage(t *testing.T) protocol.Message {
	a := assert.New(t)
	sms := NexmoSms{
		To:   "toNumber",
		From: "FromNUmber",
		Text: "body",
	}
	d, err := json.Marshal(&sms)
	if err != nil {
		a.FailNow("Could not obtain message")
	}

	msg := protocol.Message{
		Path:          protocol.Path(SMSDefaultTopic),
		UserID:        "samsa",
		ApplicationID: "sms",
		ID:            uint64(4),
		Body:          d,
	}
	return msg
}

func createNexmoSender(t *testing.T) Sender {
	a := assert.New(t)
	nexmoSender, err := NewNexmoSender(KEY, SECRET)
	if err != nil {
		a.FailNow("Nexmo sender could not be created.")
	}
	return nexmoSender
}

func noResponseFromNexmoHandler(t *testing.T, noOfReq *int)  http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a:= assert.New(t)
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)

		var sms NexmoSms
		json.Unmarshal(body,&sms)
		a.Equal("body",sms.Text)
		*noOfReq--
	}
}

func dummyNexmoEndpointWithHandlerFunc(t *testing.T,expectedRequestNo *int, port string, handler func(t *testing.T,no *int)http.HandlerFunc ) {
	http.HandleFunc("/", handler(t,expectedRequestNo) )
	http.ListenAndServe(port, nil)
}
