// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mattermost/mattermost-plugin-msoffice/server/remote (interfaces: Client)

// Package mock_remote is a generated GoMock package.
package mock_remote

import (
	gomock "github.com/golang/mock/gomock"
	remote "github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	reflect "reflect"
	time "time"
)

// MockClient is a mock of Client interface
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// AcceptUserEvent mocks base method
func (m *MockClient) AcceptUserEvent(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AcceptUserEvent", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AcceptUserEvent indicates an expected call of AcceptUserEvent
func (mr *MockClientMockRecorder) AcceptUserEvent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AcceptUserEvent", reflect.TypeOf((*MockClient)(nil).AcceptUserEvent), arg0, arg1)
}

// Call mocks base method
func (m *MockClient) Call(arg0, arg1, arg2 string, arg3, arg4 interface{}) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Call", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Call indicates an expected call of Call
func (mr *MockClientMockRecorder) Call(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Call", reflect.TypeOf((*MockClient)(nil).Call), arg0, arg1, arg2, arg3, arg4)
}

// CreateCalendar mocks base method
func (m *MockClient) CreateCalendar(arg0 *remote.Calendar) (*remote.Calendar, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCalendar", arg0)
	ret0, _ := ret[0].(*remote.Calendar)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateCalendar indicates an expected call of CreateCalendar
func (mr *MockClientMockRecorder) CreateCalendar(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCalendar", reflect.TypeOf((*MockClient)(nil).CreateCalendar), arg0)
}

// CreateEvent mocks base method
func (m *MockClient) CreateEvent(arg0 *remote.Event) (*remote.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateEvent", arg0)
	ret0, _ := ret[0].(*remote.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateEvent indicates an expected call of CreateEvent
func (mr *MockClientMockRecorder) CreateEvent(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateEvent", reflect.TypeOf((*MockClient)(nil).CreateEvent), arg0)
}

// CreateSubscription mocks base method
func (m *MockClient) CreateSubscription(arg0 string) (*remote.Subscription, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSubscription", arg0)
	ret0, _ := ret[0].(*remote.Subscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateSubscription indicates an expected call of CreateSubscription
func (mr *MockClientMockRecorder) CreateSubscription(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSubscription", reflect.TypeOf((*MockClient)(nil).CreateSubscription), arg0)
}

// DeclineUserEvent mocks base method
func (m *MockClient) DeclineUserEvent(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeclineUserEvent", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeclineUserEvent indicates an expected call of DeclineUserEvent
func (mr *MockClientMockRecorder) DeclineUserEvent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeclineUserEvent", reflect.TypeOf((*MockClient)(nil).DeclineUserEvent), arg0, arg1)
}

// DeleteCalendar mocks base method
func (m *MockClient) DeleteCalendar(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCalendar", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteCalendar indicates an expected call of DeleteCalendar
func (mr *MockClientMockRecorder) DeleteCalendar(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCalendar", reflect.TypeOf((*MockClient)(nil).DeleteCalendar), arg0)
}

// DeleteSubscription mocks base method
func (m *MockClient) DeleteSubscription(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSubscription", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSubscription indicates an expected call of DeleteSubscription
func (mr *MockClientMockRecorder) DeleteSubscription(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSubscription", reflect.TypeOf((*MockClient)(nil).DeleteSubscription), arg0)
}

// FindMeetingTimes mocks base method
func (m *MockClient) FindMeetingTimes(arg0 *remote.FindMeetingTimesParameters) (*remote.MeetingTimeSuggestionResults, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindMeetingTimes", arg0)
	ret0, _ := ret[0].(*remote.MeetingTimeSuggestionResults)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindMeetingTimes indicates an expected call of FindMeetingTimes
func (mr *MockClientMockRecorder) FindMeetingTimes(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindMeetingTimes", reflect.TypeOf((*MockClient)(nil).FindMeetingTimes), arg0)
}

// GetMe mocks base method
func (m *MockClient) GetMe() (*remote.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMe")
	ret0, _ := ret[0].(*remote.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMe indicates an expected call of GetMe
func (mr *MockClientMockRecorder) GetMe() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMe", reflect.TypeOf((*MockClient)(nil).GetMe))
}

// GetNotificationData mocks base method
func (m *MockClient) GetNotificationData(arg0 *remote.Notification) (*remote.Notification, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNotificationData", arg0)
	ret0, _ := ret[0].(*remote.Notification)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNotificationData indicates an expected call of GetNotificationData
func (mr *MockClientMockRecorder) GetNotificationData(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNotificationData", reflect.TypeOf((*MockClient)(nil).GetNotificationData), arg0)
}

// GetSchedule mocks base method
func (m *MockClient) GetSchedule(arg0 []string, arg1, arg2 *remote.DateTime, arg3 int) ([]*remote.ScheduleInformation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSchedule", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*remote.ScheduleInformation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSchedule indicates an expected call of GetSchedule
func (mr *MockClientMockRecorder) GetSchedule(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSchedule", reflect.TypeOf((*MockClient)(nil).GetSchedule), arg0, arg1, arg2, arg3)
}

// GetUserCalendars mocks base method
func (m *MockClient) GetUserCalendars(arg0 string) ([]*remote.Calendar, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserCalendars", arg0)
	ret0, _ := ret[0].([]*remote.Calendar)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserCalendars indicates an expected call of GetUserCalendars
func (mr *MockClientMockRecorder) GetUserCalendars(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserCalendars", reflect.TypeOf((*MockClient)(nil).GetUserCalendars), arg0)
}

// GetUserDefaultCalendarView mocks base method
func (m *MockClient) GetUserDefaultCalendarView(arg0 string, arg1, arg2 time.Time) ([]*remote.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserDefaultCalendarView", arg0, arg1, arg2)
	ret0, _ := ret[0].([]*remote.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserDefaultCalendarView indicates an expected call of GetUserDefaultCalendarView
func (mr *MockClientMockRecorder) GetUserDefaultCalendarView(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserDefaultCalendarView", reflect.TypeOf((*MockClient)(nil).GetUserDefaultCalendarView), arg0, arg1, arg2)
}

// GetUserEvent mocks base method
func (m *MockClient) GetUserEvent(arg0, arg1 string) (*remote.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserEvent", arg0, arg1)
	ret0, _ := ret[0].(*remote.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserEvent indicates an expected call of GetUserEvent
func (mr *MockClientMockRecorder) GetUserEvent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserEvent", reflect.TypeOf((*MockClient)(nil).GetUserEvent), arg0, arg1)
}

// ListSubscriptions mocks base method
func (m *MockClient) ListSubscriptions() ([]*remote.Subscription, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListSubscriptions")
	ret0, _ := ret[0].([]*remote.Subscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListSubscriptions indicates an expected call of ListSubscriptions
func (mr *MockClientMockRecorder) ListSubscriptions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListSubscriptions", reflect.TypeOf((*MockClient)(nil).ListSubscriptions))
}

// RenewSubscription mocks base method
func (m *MockClient) RenewSubscription(arg0 string) (*remote.Subscription, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RenewSubscription", arg0)
	ret0, _ := ret[0].(*remote.Subscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RenewSubscription indicates an expected call of RenewSubscription
func (mr *MockClientMockRecorder) RenewSubscription(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RenewSubscription", reflect.TypeOf((*MockClient)(nil).RenewSubscription), arg0)
}

// TentativelyAcceptUserEvent mocks base method
func (m *MockClient) TentativelyAcceptUserEvent(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TentativelyAcceptUserEvent", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// TentativelyAcceptUserEvent indicates an expected call of TentativelyAcceptUserEvent
func (mr *MockClientMockRecorder) TentativelyAcceptUserEvent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TentativelyAcceptUserEvent", reflect.TypeOf((*MockClient)(nil).TentativelyAcceptUserEvent), arg0, arg1)
}
