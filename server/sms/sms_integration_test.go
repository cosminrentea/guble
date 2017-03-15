package sms

import (
	"encoding/json"
	"github.com/cosminrentea/gobbler/protocol"
	"github.com/cosminrentea/gobbler/server/auth"
	"github.com/cosminrentea/gobbler/server/kvstore"
	"github.com/cosminrentea/gobbler/server/router"
	"github.com/cosminrentea/gobbler/server/store/dummystore"
	"github.com/cosminrentea/gobbler/testutil"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	//"os"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
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
	a.Equal(ErrRetryFailed,err)

	a.Equal(0, expectedRequestNo, "Three retries should be made by sender.")
	time.Sleep(timeInterval)
}

func Test_NexmoHTTPError(t *testing.T) {
	defer testutil.EnableDebugForMethod()()
	a := assert.New(t)

	port := createRandomPort(7000, 9000)
	URL = "http://127.0.0.1" + port

	expectedRequestNo := 3
	go dummyNexmoEndpointWithHandlerFunc(t, &expectedRequestNo, port, noResponseFromNexmoHandler)

	sender := createNexmoSender(t)
	config := createConfig()
	kvStore, f := createKVStore(t, "/guble_sms_nexmo_error")
	defer os.Remove(f)
	msgStore := dummystore.New(kvStore)
	accessManager := auth.NewAllowAllAccessManager(true)

	router := router.New(accessManager, msgStore, kvStore, nil)

	gw, err := New(router, sender, config)
	a.NoError(err)
	err = gw.Start()
	a.NoError(err)

	msg := encodeProtocolMessage(t, 2)
	err = gw.route.Deliver(&msg, false)
	a.NoError(err)
	time.Sleep(4 * timeInterval)
	a.Equal(0, expectedRequestNo, "Three retries should be made by sender.")
	a.Equal(msg.ID, gw.LastIDSent, "Retry failed.Last id  sent should be msgId")

	err = gw.Stop()
	time.Sleep(timeInterval)
	a.NoError(err)
}

func Test_NexmoInvalidSenderError(t *testing.T) {
	defer testutil.EnableDebugForMethod()()
	a := assert.New(t)

	port := createRandomPort(7000, 9000)
	URL = "http://127.0.0.1" + port

	expectedRequestNo := 1
	go dummyNexmoEndpointWithHandlerFunc(t, &expectedRequestNo, port, invalidSenderNexmoHandler)

	sender := createNexmoSender(t)
	config := createConfig()
	kvStore, f := createKVStore(t, "/guble_sms_nexmo_invalid_sender_error")
	defer os.Remove(f)
	msgStore := dummystore.New(kvStore)
	accessManager := auth.NewAllowAllAccessManager(true)

	router := router.New(accessManager, msgStore, kvStore, nil)

	gw, err := New(router, sender, config)
	a.NoError(err)
	err = gw.Start()
	a.NoError(err)

	msg := encodeProtocolMessage(t, 2)
	err = gw.route.Deliver(&msg, false)
	a.NoError(err)
	//time.Sleep(4 * timeInterval)
	time.Sleep(timeInterval)
	a.Equal(0, expectedRequestNo, "Only one try should be made by sender.")
	a.Equal(msg.ID, gw.LastIDSent, "No Retry needed.Last id  sent should be msgId")

	err = gw.Stop()
	time.Sleep(timeInterval)
	a.NoError(err)
}


func Test_NexmoResponseCodeError(t *testing.T) {
	defer testutil.EnableDebugForMethod()()
	a := assert.New(t)

	port := createRandomPort(7000, 9000)
	URL = "http://127.0.0.1" + port

	expectedRequestNo := 3
	go dummyNexmoEndpointWithHandlerFunc(t, &expectedRequestNo, port, responseInternalErrorNexmoHandler)

	sender := createNexmoSender(t)
	config := createConfig()
	kvStore, f := createKVStore(t, "/guble_sms_nexmo_invalid_sender_error")
	defer os.Remove(f)
	msgStore := dummystore.New(kvStore)
	accessManager := auth.NewAllowAllAccessManager(true)

	router := router.New(accessManager, msgStore, kvStore, nil)

	gw, err := New(router, sender, config)
	a.NoError(err)
	err = gw.Start()
	a.NoError(err)

	msg := encodeProtocolMessage(t, 2)
	err = gw.route.Deliver(&msg, false)
	a.NoError(err)
	time.Sleep(5 * timeInterval)
	a.Equal(0, expectedRequestNo, "Only one try should be made by sender.")
	a.Equal(msg.ID, gw.LastIDSent, "No Retry needed.Last id  sent should be msgId")

	err = gw.Stop()
	time.Sleep(timeInterval)
	a.NoError(err)
}


func createConfig() Config {
	topic := "/sms"
	worker := 1
	intervalMetrics := true
	return Config{
		Workers:  &worker,
		SMSTopic: &topic,
		Name:     "test_gateway",
		Schema:   SMSSchema,

		IntervalMetrics: &intervalMetrics,
	}

}

func createKVStore(t *testing.T, filename string) (kvstore.KVStore, string) {
	a := assert.New(t)
	f := tempFilename(filename)
	//defer os.Remove(f)

	//create a KVStore with sqlite3
	kvStore := kvstore.NewSqliteKVStore(f, true)
	err := kvStore.Open()
	if err != nil {
		a.FailNow("KVStore could not be opened.")
	}
	a.NotNil(kvStore)
	return kvStore, f
}

func encodeProtocolMessage(t *testing.T, ID int) protocol.Message {
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
		ID:            uint64(ID),
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

func noResponseFromNexmoHandler(t *testing.T, noOfReq *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := assert.New(t)
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)

		var sms NexmoSms
		json.Unmarshal(body, &sms)
		a.Equal("body", sms.Text)
		*noOfReq--
	}
}

func invalidSenderNexmoHandler(t *testing.T, noOfReq *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := assert.New(t)
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)

		var sms NexmoSms
		json.Unmarshal(body, &sms)
		a.Equal("body", sms.Text)
		*noOfReq--

		nexmoResponse := composeNexmoMessageResponse(sms, ResponseInvalidSenderAddress)

		jData, err := json.Marshal(nexmoResponse)
		if err != nil {
			a.FailNow("Nexmo Response encoding failed.")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jData)

	}
}

func responseInternalErrorNexmoHandler(t *testing.T, noOfReq *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := assert.New(t)
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)

		var sms NexmoSms
		json.Unmarshal(body, &sms)
		a.Equal("body", sms.Text)
		*noOfReq--

		nexmoResponse := composeNexmoMessageResponse(sms, ResponseInternalError)

		jData, err := json.Marshal(nexmoResponse)
		if err != nil {
			a.FailNow("Nexmo Response encoding failed.")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jData)

	}
}

func composeNexmoMessageResponse(sms NexmoSms, code ResponseCode) *NexmoMessageResponse {
	nexmoResponse := new(NexmoMessageResponse)
	nexmoResponse.MessageCount = 1
	response := NexmoMessageReport{
		Status:           code,
		MessageID:        "msgID",
		To:               sms.To,
		ClientReference:  "",
		RemainingBalance: "2",
		MessagePrice:     "0.005",
		Network:          "TELEKOM",
		ErrorText:        "",
	}
	nexmoResponse.Messages = []NexmoMessageReport{response}
	return nexmoResponse
}

func dummyNexmoEndpointWithHandlerFunc(t *testing.T, expectedRequestNo *int, port string, handler func(t *testing.T, no *int) http.HandlerFunc) {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", handler(t, expectedRequestNo))
	fmt.Println(port)
	http.ListenAndServe(port, serveMux)
}

func createRandomPort(min, max int) string {
	rand.Seed(time.Now().UnixNano())
	port := rand.Intn(max-min) + min
	return fmt.Sprintf(":%d", port)
}
