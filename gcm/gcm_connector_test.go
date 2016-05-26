package gcm

import (
	"github.com/golang/mock/gomock"
	"github.com/smancke/guble/protocol"
	"github.com/smancke/guble/server"
	"github.com/smancke/guble/store"
	"github.com/stretchr/testify/assert"

	"fmt"
	"github.com/alexjlockwood/gcm"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

var ctrl *gomock.Controller
var correctGcmResponseMessageJSON = `
{
   "multicast_id":3,
   "succes":1,
   "failure":0,
   "canonicals_ids":5,
   "results":[
      {
         "message_id":"da",
         "registration_id":"rId",
         "error":""
      }
   ]
}`

// mock a https round tripper in order to not send the test reques gcm.
type RoundTripperFunc func(req *http.Request) *http.Response

func (rt RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	protocol.Debug("am intrat aicishea")
	return rt(req), nil
}

func createSender(rt RoundTripperFunc) *gcm.Sender {
	httpClient := &http.Client{Transport: rt}
	return &gcm.Sender{ApiKey: "124", Http: httpClient}

}

func composeHTTPResponse(req *http.Request, httpStatusCode int, messageBodyAsJSON string) *http.Response {
	resp := &http.Response{
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       ioutil.NopCloser(strings.NewReader(messageBodyAsJSON)),
		Request:    req,
		StatusCode: httpStatusCode,
	}
	resp.Header.Add("Content-Type", "application/json")
	return resp
}

func TestServeHTTPSuccess(t *testing.T) {
	defer initCtrl(t)()

	a := assert.New(t)

	// given:  a rest api with a message sink
	routerMock := NewMockRouter(ctrl)

	kvStore := store.NewMemoryKVStore()
	routerMock.EXPECT().KVStore().Return(kvStore, nil)

	routerMock.EXPECT().Subscribe(gomock.Any()).Do(func(route *server.Route) {
		a.Equal("/notifications", string(route.Path))
		a.Equal("marvin", route.UserID)
		a.Equal("gcmId123", route.ApplicationID)
	})

	gcm, err := NewGCMConnector(routerMock, "/gcm/", "testApi")
	a.Nil(err)

	url, _ := url.Parse("http://localhost/gcm/marvin/gcmId123/subscribe/notifications")
	// and a http context
	req := &http.Request{URL: url, Method: "POST"}
	w := httptest.NewRecorder()

	// when: I POST a message
	gcm.ServeHTTP(w, req)

	// the the result
	a.Equal("registered: /notifications\n", string(w.Body.Bytes()))
}

func TestServeHTTPWithErrorCases(t *testing.T) {
	defer initCtrl(t)()

	a := assert.New(t)

	// given:  a rest api with a message sink
	routerMock := NewMockRouter(ctrl)

	kvStore := store.NewMemoryKVStore()
	routerMock.EXPECT().KVStore().Return(kvStore, nil)

	gcm, err := NewGCMConnector(routerMock, "/gcm/", "testApi")
	a.Nil(err)

	url, _ := url.Parse("http://localhost/gcm/marvin/gcmId123/subscribe/notifications")
	// and a http context
	req := &http.Request{URL: url, Method: "GET"}
	w := httptest.NewRecorder()

	// do a GET instead of POST
	gcm.ServeHTTP(w, req)

	// check the result
	a.Equal("Permission Denied\n", string(w.Body.Bytes()))
	a.Equal(w.Code, 405)

	// send a new request with wrong parameters encoding
	req.Method = "POST"
	req.URL, _ = url.Parse("http://localhost/gcm/marvin/gcmId123/subscribe3/notifications")

	w2 := httptest.NewRecorder()
	gcm.ServeHTTP(w2, req)

	a.Equal("Invalid Parameters in request\n", string(w2.Body.Bytes()))
	a.Equal(w2.Code, http.StatusBadRequest)

}

func TestSaveAndLoadSubscriptions(t *testing.T) {
	defer initCtrl(t)()
	defer enableDebugForMethod()()
	a := assert.New(t)

	// given: some test routes
	testRoutes := map[string]bool{
		"marvin:/foo:1234": true,
		"zappod:/bar:1212": true,
		"athur:/erde:42":   true,
	}

	routerMock := NewMockRouter(ctrl)

	kvStore := store.NewMemoryKVStore()
	routerMock.EXPECT().KVStore().Return(kvStore, nil)

	routerMock.EXPECT().Subscribe(gomock.Any()).Do(func(route *server.Route) {
		// delete the route from the map, if we got it in the test
		delete(testRoutes, fmt.Sprintf("%v:%v:%v", route.UserID, route.Path, route.ApplicationID))
	}).AnyTimes()

	gcm, err := NewGCMConnector(routerMock, "/gcm/", "testApi")
	a.Nil(err)

	// when: we save the routes
	for k := range testRoutes {
		splitedKey := strings.SplitN(k, ":", 3)
		userid := splitedKey[0]
		topic := splitedKey[1]
		gcmid := splitedKey[2]
		gcm.saveSubscription(userid, topic, gcmid)
	}

	// and reload the routes
	gcm.loadSubscriptions()

	time.Sleep(time.Millisecond * 100)

	// than: all expected subscriptions were called
	a.Equal(0, len(testRoutes))
}

func TestRemoveTailingSlash(t *testing.T) {
	assert.Equal(t, "/foo", removeTrailingSlash("/foo/"))
	assert.Equal(t, "/foo", removeTrailingSlash("/foo"))
}

func TestGCMConnector_parseParams(t *testing.T) {
	defer initCtrl(t)()

	assert := assert.New(t)
	routerMock := NewMockRouter(ctrl)
	kvStore := store.NewMemoryKVStore()
	routerMock.EXPECT().KVStore().Return(kvStore, nil)

	gcm, err := NewGCMConnector(routerMock, "/gcm/", "testApi")
	assert.Nil(err)

	testCases := []struct {
		urlPath, userID, gcmID, topic, err string
	}{
		{"/gcm/marvin/gcmId123/subscribe/notifications", "marvin", "gcmId123", "/notifications", ""},
		{"/gcm2/marvin/gcmId123/subscribe/notifications", "", "", "", "Gcm request is not starting with gcm prefix"},
		{"/gcm/marvin/gcmId123/subscrib2e/notifications", "", "", "", "Gcm request third param is not subscribe"},
		{"/gcm/marvin/gcmId123subscribenotifications", "", "", "", "Gcm request has wrong number of params"},
		{"/gcm/marvin/gcmId123/subscribe/notifications/alert/", "marvin", "gcmId123", "/notifications/alert", ""},
	}

	for i, c := range testCases {
		userID, gcmID, topic, err := gcm.parseParams(c.urlPath)

		//if error message is present check only the error
		if c.err != "" {
			assert.NotNil(err)
			assert.EqualError(err, c.err, fmt.Sprintf("Failed on testcase no=%d", i))
		} else {
			assert.Equal(userID, c.userID, fmt.Sprintf("Failed on testcase no=%d", i))
			assert.Equal(gcmID, c.gcmID, fmt.Sprintf("Failed on testcase no=%d", i))
			assert.Equal(topic, c.topic, fmt.Sprintf("Failed on testcase no=%d", i))
			assert.Nil(err, fmt.Sprintf("Failed on testcase no=%d", i))
		}

	}

}

func TestGCMConnector_GetPrefix(t *testing.T) {
	defer initCtrl(t)()

	assert := assert.New(t)
	routerMock := NewMockRouter(ctrl)
	kvStore := store.NewMemoryKVStore()
	routerMock.EXPECT().KVStore().Return(kvStore, nil)

	gcm, err := NewGCMConnector(routerMock, "/gcm/", "testApi")
	assert.Nil(err)
	assert.Equal(gcm.GetPrefix(), "/gcm/")
}

func TestGCMConnector_Stop(t *testing.T) {
	defer initCtrl(t)()

	assert := assert.New(t)
	routerMock := NewMockRouter(ctrl)
	kvStore := store.NewMemoryKVStore()
	routerMock.EXPECT().KVStore().Return(kvStore, nil)

	gcm, err := NewGCMConnector(routerMock, "/gcm/", "testApi")
	assert.Nil(err)

	err = gcm.Stop()
	assert.Nil(err)
	assert.Equal(len(gcm.stopChan), 1, "StopChan")

}

func TestGcmConnector_StartWithMessageSending(t *testing.T) {
	defer initCtrl(t)()

	assert := assert.New(t)
	routerMock := NewMockRouter(ctrl)
	routerMock.EXPECT().Subscribe(gomock.Any()).Do(func(route *server.Route) {
		assert.Equal("/gcm/broadcast", string(route.Path))
		assert.Equal("gcm_connector", route.UserID)
		assert.Equal("gcm_connector", route.ApplicationID)
	})

	kvStore := store.NewMemoryKVStore()
	routerMock.EXPECT().KVStore().Return(kvStore, nil)

	gcm, err := NewGCMConnector(routerMock, "/gcm/", "testApi")
	assert.Nil(err)

	err = gcm.Start()
	assert.Nil(err)

	done := make(chan bool, 1)
	mockSender := createSender(RoundTripperFunc(func(req *http.Request) *http.Response {
		// signal the ending of processing
		defer func() {
			done <- true
		}()

		return composeHTTPResponse(req, http.StatusOK, correctGcmResponseMessageJSON)
	}))
	gcm.sender = mockSender

	// put a broadcast message with no recipients and expect to be dropped by
	broadcastMsgWithNoRecipients := server.MsgAndRoute{Message: &protocol.Message{ID: uint64(4), Body: []byte("{id:id}"), Time: 1405544146, Path: "/gcm/broadcast"}}
	gcm.channelFromRouter <- broadcastMsgWithNoRecipients
	time.Sleep(time.Millisecond * 1000)
	// expect that the HTTP Dummy Server to not handle any requests

	// put a dummy gcm message with minimum information
	msgWithNoRecipients := server.MsgAndRoute{Message: &protocol.Message{ID: uint64(4), Body: []byte("{id:id}"), Time: 1405544146, Path: "/gcm/marvin/gcm124/subscribe/stuff"}, Route: &server.Route{ApplicationID: "id"}}
	gcm.channelFromRouter <- msgWithNoRecipients
	// expect that the Http Server to give us a malformed message
	<-done

	//wait a couple of seconds and  Stop the GcmConnector
	time.Sleep(time.Millisecond * 2000)
	err = gcm.Stop()
	assert.Nil(err)
}

func initCtrl(t *testing.T) func() {
	ctrl = gomock.NewController(t)
	return func() {
		ctrl.Finish()
	}
}

func enableDebugForMethod() func() {
	reset := protocol.LogLevel
	protocol.LogLevel = protocol.LEVEL_DEBUG
	return func() {
		protocol.LogLevel = reset
	}
}
