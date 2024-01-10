// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote (interfaces: Remote)

// Package mock_remote is a generated GoMock package.
package mock_remote

import (
	context "context"
	http "net/http"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	config "github.com/mattermost/mattermost-plugin-mscalendar/calendar/config"
	remote "github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
	oauth2 "golang.org/x/oauth2"
)

// MockRemote is a mock of Remote interface.
type MockRemote struct {
	ctrl     *gomock.Controller
	recorder *MockRemoteMockRecorder
}

// MockRemoteMockRecorder is the mock recorder for MockRemote.
type MockRemoteMockRecorder struct {
	mock *MockRemote
}

// NewMockRemote creates a new mock instance.
func NewMockRemote(ctrl *gomock.Controller) *MockRemote {
	mock := &MockRemote{ctrl: ctrl}
	mock.recorder = &MockRemoteMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRemote) EXPECT() *MockRemoteMockRecorder {
	return m.recorder
}

// CheckConfiguration mocks base method.
func (m *MockRemote) CheckConfiguration(arg0 config.StoredConfig) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckConfiguration", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CheckConfiguration indicates an expected call of CheckConfiguration.
func (mr *MockRemoteMockRecorder) CheckConfiguration(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckConfiguration", reflect.TypeOf((*MockRemote)(nil).CheckConfiguration), arg0)
}

// HandleWebhook mocks base method.
func (m *MockRemote) HandleWebhook(arg0 http.ResponseWriter, arg1 *http.Request) []*remote.Notification {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HandleWebhook", arg0, arg1)
	ret0, _ := ret[0].([]*remote.Notification)
	return ret0
}

// HandleWebhook indicates an expected call of HandleWebhook.
func (mr *MockRemoteMockRecorder) HandleWebhook(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleWebhook", reflect.TypeOf((*MockRemote)(nil).HandleWebhook), arg0, arg1)
}

// MakeClient mocks base method.
func (m *MockRemote) MakeClient(arg0 context.Context, arg1 *oauth2.Token) remote.Client {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MakeClient", arg0, arg1)
	ret0, _ := ret[0].(remote.Client)
	return ret0
}

// MakeClient indicates an expected call of MakeClient.
func (mr *MockRemoteMockRecorder) MakeClient(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MakeClient", reflect.TypeOf((*MockRemote)(nil).MakeClient), arg0, arg1)
}

// MakeSuperuserClient mocks base method.
func (m *MockRemote) MakeSuperuserClient(arg0 context.Context) (remote.Client, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MakeSuperuserClient", arg0)
	ret0, _ := ret[0].(remote.Client)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MakeSuperuserClient indicates an expected call of MakeSuperuserClient.
func (mr *MockRemoteMockRecorder) MakeSuperuserClient(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MakeSuperuserClient", reflect.TypeOf((*MockRemote)(nil).MakeSuperuserClient), arg0)
}

// NewOAuth2Config mocks base method.
func (m *MockRemote) NewOAuth2Config() *oauth2.Config {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewOAuth2Config")
	ret0, _ := ret[0].(*oauth2.Config)
	return ret0
}

// NewOAuth2Config indicates an expected call of NewOAuth2Config.
func (mr *MockRemoteMockRecorder) NewOAuth2Config() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewOAuth2Config", reflect.TypeOf((*MockRemote)(nil).NewOAuth2Config))
}
