// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mattermost/mattermost-plugin-msoffice/server/api (interfaces: Calendar)

// Package mock_api is a generated GoMock package.
package mock_api

import (
	gomock "github.com/golang/mock/gomock"
	remote "github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	reflect "reflect"
	time "time"
)

// MockCalendar is a mock of Calendar interface
type MockCalendar struct {
	ctrl     *gomock.Controller
	recorder *MockCalendarMockRecorder
}

// MockCalendarMockRecorder is the mock recorder for MockCalendar
type MockCalendarMockRecorder struct {
	mock *MockCalendar
}

// NewMockCalendar creates a new mock instance
func NewMockCalendar(ctrl *gomock.Controller) *MockCalendar {
	mock := &MockCalendar{ctrl: ctrl}
	mock.recorder = &MockCalendarMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCalendar) EXPECT() *MockCalendarMockRecorder {
	return m.recorder
}

// CreateCalendar mocks base method
func (m *MockCalendar) CreateCalendar(arg0 *remote.Calendar) (*remote.Calendar, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCalendar", arg0)
	ret0, _ := ret[0].(*remote.Calendar)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateCalendar indicates an expected call of CreateCalendar
func (mr *MockCalendarMockRecorder) CreateCalendar(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCalendar", reflect.TypeOf((*MockCalendar)(nil).CreateCalendar), arg0)
}

// CreateEvent mocks base method
func (m *MockCalendar) CreateEvent(arg0 *remote.Event) (*remote.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateEvent", arg0)
	ret0, _ := ret[0].(*remote.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateEvent indicates an expected call of CreateEvent
func (mr *MockCalendarMockRecorder) CreateEvent(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateEvent", reflect.TypeOf((*MockCalendar)(nil).CreateEvent), arg0)
}

// DeleteCalendar mocks base method
func (m *MockCalendar) DeleteCalendar(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCalendar", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteCalendar indicates an expected call of DeleteCalendar
func (mr *MockCalendarMockRecorder) DeleteCalendar(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCalendar", reflect.TypeOf((*MockCalendar)(nil).DeleteCalendar), arg0)
}

// FindMeetingTimes mocks base method
func (m *MockCalendar) FindMeetingTimes(arg0 *remote.FindMeetingTimesParameters) (*remote.MeetingTimeSuggestionResults, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindMeetingTimes", arg0)
	ret0, _ := ret[0].(*remote.MeetingTimeSuggestionResults)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindMeetingTimes indicates an expected call of FindMeetingTimes
func (mr *MockCalendarMockRecorder) FindMeetingTimes(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindMeetingTimes", reflect.TypeOf((*MockCalendar)(nil).FindMeetingTimes), arg0)
}

// GetUserCalendars mocks base method
func (m *MockCalendar) GetUserCalendars(arg0 string) ([]*remote.Calendar, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserCalendars", arg0)
	ret0, _ := ret[0].([]*remote.Calendar)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserCalendars indicates an expected call of GetUserCalendars
func (mr *MockCalendarMockRecorder) GetUserCalendars(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserCalendars", reflect.TypeOf((*MockCalendar)(nil).GetUserCalendars), arg0)
}

// ViewCalendar mocks base method
func (m *MockCalendar) ViewCalendar(arg0, arg1 time.Time) ([]*remote.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ViewCalendar", arg0, arg1)
	ret0, _ := ret[0].([]*remote.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ViewCalendar indicates an expected call of ViewCalendar
func (mr *MockCalendarMockRecorder) ViewCalendar(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ViewCalendar", reflect.TypeOf((*MockCalendar)(nil).ViewCalendar), arg0, arg1)
}
