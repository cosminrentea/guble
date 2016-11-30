// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/smancke/guble/server/connector (interfaces: Connector,Sender,ResponseHandler,Manager,Queue,Request,Subscriber)

package connector

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/smancke/guble/protocol"

	"github.com/smancke/guble/server/router"
	"net/http"
)

// Mock of Connector interface
type MockConnector struct {
	ctrl     *gomock.Controller
	recorder *_MockConnectorRecorder
}

// Recorder for MockConnector (not exported)
type _MockConnectorRecorder struct {
	mock *MockConnector
}

func NewMockConnector(ctrl *gomock.Controller) *MockConnector {
	mock := &MockConnector{ctrl: ctrl}
	mock.recorder = &_MockConnectorRecorder{mock}
	return mock
}

func (_m *MockConnector) EXPECT() *_MockConnectorRecorder {
	return _m.recorder
}

func (_m *MockConnector) Context() context.Context {
	ret := _m.ctrl.Call(_m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

func (_mr *_MockConnectorRecorder) Context() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Context")
}

func (_m *MockConnector) GetPrefix() string {
	ret := _m.ctrl.Call(_m, "GetPrefix")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockConnectorRecorder) GetPrefix() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetPrefix")
}

func (_m *MockConnector) Manager() Manager {
	ret := _m.ctrl.Call(_m, "Manager")
	ret0, _ := ret[0].(Manager)
	return ret0
}

func (_mr *_MockConnectorRecorder) Manager() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Manager")
}

func (_m *MockConnector) ResponseHandler() ResponseHandler {
	ret := _m.ctrl.Call(_m, "ResponseHandler")
	ret0, _ := ret[0].(ResponseHandler)
	return ret0
}

func (_mr *_MockConnectorRecorder) ResponseHandler() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ResponseHandler")
}

func (_m *MockConnector) Run(_param0 Subscriber) {
	_m.ctrl.Call(_m, "Run", _param0)
}

func (_mr *_MockConnectorRecorder) Run(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Run", arg0)
}

func (_m *MockConnector) Sender() Sender {
	ret := _m.ctrl.Call(_m, "Sender")
	ret0, _ := ret[0].(Sender)
	return ret0
}

func (_mr *_MockConnectorRecorder) Sender() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Sender")
}

func (_m *MockConnector) ServeHTTP(_param0 http.ResponseWriter, _param1 *http.Request) {
	_m.ctrl.Call(_m, "ServeHTTP", _param0, _param1)
}

func (_mr *_MockConnectorRecorder) ServeHTTP(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ServeHTTP", arg0, arg1)
}

func (_m *MockConnector) SetResponseHandler(_param0 ResponseHandler) {
	_m.ctrl.Call(_m, "SetResponseHandler", _param0)
}

func (_mr *_MockConnectorRecorder) SetResponseHandler(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetResponseHandler", arg0)
}

func (_m *MockConnector) SetSender(_param0 Sender) {
	_m.ctrl.Call(_m, "SetSender", _param0)
}

func (_mr *_MockConnectorRecorder) SetSender(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetSender", arg0)
}

func (_m *MockConnector) Start() error {
	ret := _m.ctrl.Call(_m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockConnectorRecorder) Start() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Start")
}

func (_m *MockConnector) Stop() error {
	ret := _m.ctrl.Call(_m, "Stop")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockConnectorRecorder) Stop() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Stop")
}

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

func (_m *MockSender) Send(_param0 Request) (interface{}, error) {
	ret := _m.ctrl.Call(_m, "Send", _param0)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockSenderRecorder) Send(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Send", arg0)
}

// Mock of ResponseHandler interface
type MockResponseHandler struct {
	ctrl     *gomock.Controller
	recorder *_MockResponseHandlerRecorder
}

// Recorder for MockResponseHandler (not exported)
type _MockResponseHandlerRecorder struct {
	mock *MockResponseHandler
}

func NewMockResponseHandler(ctrl *gomock.Controller) *MockResponseHandler {
	mock := &MockResponseHandler{ctrl: ctrl}
	mock.recorder = &_MockResponseHandlerRecorder{mock}
	return mock
}

func (_m *MockResponseHandler) EXPECT() *_MockResponseHandlerRecorder {
	return _m.recorder
}

func (_m *MockResponseHandler) HandleResponse(_param0 Request, _param1 interface{}, _param2 *Metadata, _param3 error) error {
	ret := _m.ctrl.Call(_m, "HandleResponse", _param0, _param1, _param2, _param3)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockResponseHandlerRecorder) HandleResponse(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "HandleResponse", arg0, arg1, arg2, arg3)
}

// Mock of Manager interface
type MockManager struct {
	ctrl     *gomock.Controller
	recorder *_MockManagerRecorder
}

// Recorder for MockManager (not exported)
type _MockManagerRecorder struct {
	mock *MockManager
}

func NewMockManager(ctrl *gomock.Controller) *MockManager {
	mock := &MockManager{ctrl: ctrl}
	mock.recorder = &_MockManagerRecorder{mock}
	return mock
}

func (_m *MockManager) EXPECT() *_MockManagerRecorder {
	return _m.recorder
}

func (_m *MockManager) Add(_param0 Subscriber) error {
	ret := _m.ctrl.Call(_m, "Add", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockManagerRecorder) Add(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Add", arg0)
}

func (_m *MockManager) Create(_param0 protocol.Path, _param1 router.RouteParams) (Subscriber, error) {
	ret := _m.ctrl.Call(_m, "Create", _param0, _param1)
	ret0, _ := ret[0].(Subscriber)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockManagerRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Create", arg0, arg1)
}

func (_m *MockManager) Exists(_param0 string) bool {
	ret := _m.ctrl.Call(_m, "Exists", _param0)
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockManagerRecorder) Exists(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Exists", arg0)
}

func (_m *MockManager) Filter(_param0 map[string]string) []Subscriber {
	ret := _m.ctrl.Call(_m, "Filter", _param0)
	ret0, _ := ret[0].([]Subscriber)
	return ret0
}

func (_mr *_MockManagerRecorder) Filter(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Filter", arg0)
}

func (_m *MockManager) Find(_param0 string) Subscriber {
	ret := _m.ctrl.Call(_m, "Find", _param0)
	ret0, _ := ret[0].(Subscriber)
	return ret0
}

func (_mr *_MockManagerRecorder) Find(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Find", arg0)
}

func (_m *MockManager) List() []Subscriber {
	ret := _m.ctrl.Call(_m, "List")
	ret0, _ := ret[0].([]Subscriber)
	return ret0
}

func (_mr *_MockManagerRecorder) List() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "List")
}

func (_m *MockManager) Load() error {
	ret := _m.ctrl.Call(_m, "Load")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockManagerRecorder) Load() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Load")
}

func (_m *MockManager) Remove(_param0 Subscriber) error {
	ret := _m.ctrl.Call(_m, "Remove", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockManagerRecorder) Remove(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Remove", arg0)
}

func (_m *MockManager) Update(_param0 Subscriber) error {
	ret := _m.ctrl.Call(_m, "Update", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockManagerRecorder) Update(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Update", arg0)
}

// Mock of Queue interface
type MockQueue struct {
	ctrl     *gomock.Controller
	recorder *_MockQueueRecorder
}

// Recorder for MockQueue (not exported)
type _MockQueueRecorder struct {
	mock *MockQueue
}

func NewMockQueue(ctrl *gomock.Controller) *MockQueue {
	mock := &MockQueue{ctrl: ctrl}
	mock.recorder = &_MockQueueRecorder{mock}
	return mock
}

func (_m *MockQueue) EXPECT() *_MockQueueRecorder {
	return _m.recorder
}

func (_m *MockQueue) Push(_param0 Request) error {
	ret := _m.ctrl.Call(_m, "Push", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockQueueRecorder) Push(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Push", arg0)
}

func (_m *MockQueue) ResponseHandler() ResponseHandler {
	ret := _m.ctrl.Call(_m, "ResponseHandler")
	ret0, _ := ret[0].(ResponseHandler)
	return ret0
}

func (_mr *_MockQueueRecorder) ResponseHandler() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ResponseHandler")
}

func (_m *MockQueue) Sender() Sender {
	ret := _m.ctrl.Call(_m, "Sender")
	ret0, _ := ret[0].(Sender)
	return ret0
}

func (_mr *_MockQueueRecorder) Sender() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Sender")
}

func (_m *MockQueue) SetResponseHandler(_param0 ResponseHandler) {
	_m.ctrl.Call(_m, "SetResponseHandler", _param0)
}

func (_mr *_MockQueueRecorder) SetResponseHandler(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetResponseHandler", arg0)
}

func (_m *MockQueue) SetSender(_param0 Sender) {
	_m.ctrl.Call(_m, "SetSender", _param0)
}

func (_mr *_MockQueueRecorder) SetSender(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetSender", arg0)
}

func (_m *MockQueue) Start() error {
	ret := _m.ctrl.Call(_m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockQueueRecorder) Start() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Start")
}

func (_m *MockQueue) Stop() error {
	ret := _m.ctrl.Call(_m, "Stop")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockQueueRecorder) Stop() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Stop")
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

func (_m *MockRequest) Subscriber() Subscriber {
	ret := _m.ctrl.Call(_m, "Subscriber")
	ret0, _ := ret[0].(Subscriber)
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

func (_m *MockSubscriber) Loop(_param0 context.Context, _param1 Queue) error {
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
