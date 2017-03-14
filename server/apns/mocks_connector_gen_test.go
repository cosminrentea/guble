// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/cosminrentea/gobbler/server/connector (interfaces: Sender,Request,Subscriber)

package apns

import (
	context "context"
	protocol "github.com/cosminrentea/gobbler/protocol"
	connector "github.com/cosminrentea/gobbler/server/connector"
	router "github.com/cosminrentea/gobbler/server/router"
	gomock "github.com/golang/mock/gomock"
)

// Mock of Sender interface
type MockSender struct {
	ctrl     *gomock.Controller
	recorder *_MockSenderRecorder
}

// Recorder for MockSender (not exported)
type _MockSenderRecorder struct {
	mock *MockSender
}

func NewMockSender(ctrl *gomock.Controller) *MockSender {
	mock := &MockSender{ctrl: ctrl}
	mock.recorder = &_MockSenderRecorder{mock}
	return mock
}

func (_m *MockSender) EXPECT() *_MockSenderRecorder {
	return _m.recorder
}

func (_m *MockSender) Send(_param0 connector.Request) (interface{}, error) {
	ret := _m.ctrl.Call(_m, "Send", _param0)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockSenderRecorder) Send(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Send", arg0)
}

// Mock of Request interface
type MockRequest struct {
	ctrl     *gomock.Controller
	recorder *_MockRequestRecorder
}

// Recorder for MockRequest (not exported)
type _MockRequestRecorder struct {
	mock *MockRequest
}

func NewMockRequest(ctrl *gomock.Controller) *MockRequest {
	mock := &MockRequest{ctrl: ctrl}
	mock.recorder = &_MockRequestRecorder{mock}
	return mock
}

func (_m *MockRequest) EXPECT() *_MockRequestRecorder {
	return _m.recorder
}

func (_m *MockRequest) Message() *protocol.Message {
	ret := _m.ctrl.Call(_m, "Message")
	ret0, _ := ret[0].(*protocol.Message)
	return ret0
}

func (_mr *_MockRequestRecorder) Message() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Message")
}

func (_m *MockRequest) Subscriber() connector.Subscriber {
	ret := _m.ctrl.Call(_m, "Subscriber")
	ret0, _ := ret[0].(connector.Subscriber)
	return ret0
}

func (_mr *_MockRequestRecorder) Subscriber() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Subscriber")
}

// Mock of Subscriber interface
type MockSubscriber struct {
	ctrl     *gomock.Controller
	recorder *_MockSubscriberRecorder
}

// Recorder for MockSubscriber (not exported)
type _MockSubscriberRecorder struct {
	mock *MockSubscriber
}

func NewMockSubscriber(ctrl *gomock.Controller) *MockSubscriber {
	mock := &MockSubscriber{ctrl: ctrl}
	mock.recorder = &_MockSubscriberRecorder{mock}
	return mock
}

func (_m *MockSubscriber) EXPECT() *_MockSubscriberRecorder {
	return _m.recorder
}

func (_m *MockSubscriber) Cancel() {
	_m.ctrl.Call(_m, "Cancel")
}

func (_mr *_MockSubscriberRecorder) Cancel() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Cancel")
}

func (_m *MockSubscriber) Encode() ([]byte, error) {
	ret := _m.ctrl.Call(_m, "Encode")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockSubscriberRecorder) Encode() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Encode")
}

func (_m *MockSubscriber) Filter(_param0 map[string]string) bool {
	ret := _m.ctrl.Call(_m, "Filter", _param0)
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockSubscriberRecorder) Filter(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Filter", arg0)
}

func (_m *MockSubscriber) Key() string {
	ret := _m.ctrl.Call(_m, "Key")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockSubscriberRecorder) Key() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Key")
}

func (_m *MockSubscriber) Loop(_param0 context.Context, _param1 connector.Queue) error {
	ret := _m.ctrl.Call(_m, "Loop", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockSubscriberRecorder) Loop(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Loop", arg0, arg1)
}

func (_m *MockSubscriber) Reset() error {
	ret := _m.ctrl.Call(_m, "Reset")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockSubscriberRecorder) Reset() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Reset")
}

func (_m *MockSubscriber) Route() *router.Route {
	ret := _m.ctrl.Call(_m, "Route")
	ret0, _ := ret[0].(*router.Route)
	return ret0
}

func (_mr *_MockSubscriberRecorder) Route() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Route")
}

func (_m *MockSubscriber) SetLastID(_param0 uint64) {
	_m.ctrl.Call(_m, "SetLastID", _param0)
}

func (_mr *_MockSubscriberRecorder) SetLastID(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetLastID", arg0)
}
