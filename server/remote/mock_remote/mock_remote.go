// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mattermost/mattermost-plugin-msoffice/server/remote (interfaces: Remote)

// Package mock_remote is a generated GoMock package.
package mock_remote

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	remote "github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	oauth2 "golang.org/x/oauth2"
	http "net/http"
	reflect "reflect"
)

// MockRemote is a mock of Remote interface
type MockRemote struct {
	ctrl     *gomock.Controller
	recorder *MockRemoteMockRecorder
}

// MockRemoteMockRecorder is the mock recorder for MockRemote
type MockRemoteMockRecorder struct {
	mock *MockRemote
}

// NewMockRemote creates a new mock instance
func NewMockRemote(ctrl *gomock.Controller) *MockRemote {
	mock := &MockRemote{ctrl: ctrl}
	mock.recorder = &MockRemoteMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRemote) EXPECT() *MockRemoteMockRecorder {
	return m.recorder
}

// HandleEventNotification mocks base method
func (m *MockRemote) HandleEventNotification(arg0 http.ResponseWriter, arg1 *http.Request, arg2 remote.LoadSubscriptionCreatorF) []*remote.EventNotification {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HandleEventNotification", arg0, arg1, arg2)
	ret0, _ := ret[0].([]*remote.EventNotification)
	return ret0
}

// HandleEventNotification indicates an expected call of HandleEventNotification
func (mr *MockRemoteMockRecorder) HandleEventNotification(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleEventNotification", reflect.TypeOf((*MockRemote)(nil).HandleEventNotification), arg0, arg1, arg2)
}

// NewClient mocks base method
func (m *MockRemote) NewClient(arg0 context.Context, arg1 *oauth2.Token) remote.Client {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewClient", arg0, arg1)
	ret0, _ := ret[0].(remote.Client)
	return ret0
}

// NewClient indicates an expected call of NewClient
func (mr *MockRemoteMockRecorder) NewClient(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewClient", reflect.TypeOf((*MockRemote)(nil).NewClient), arg0, arg1)
}

// NewOAuth2Config mocks base method
func (m *MockRemote) NewOAuth2Config() *oauth2.Config {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewOAuth2Config")
	ret0, _ := ret[0].(*oauth2.Config)
	return ret0
}

// NewOAuth2Config indicates an expected call of NewOAuth2Config
func (mr *MockRemoteMockRecorder) NewOAuth2Config() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewOAuth2Config", reflect.TypeOf((*MockRemote)(nil).NewOAuth2Config))
}
