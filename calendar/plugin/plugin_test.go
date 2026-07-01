// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package plugin

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/httputils"
)

func TestServeHTTPRecoversFromPanic(t *testing.T) {
	mockAPI := &plugintest.API{}
	mockAPI.On("LogError", "Recovered from panic while serving HTTP request", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	defer mockAPI.AssertExpectations(t)

	handler := httputils.NewHandler()
	handler.Router.HandleFunc("/panic", func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}).Methods(http.MethodGet)

	p := &Plugin{envLock: &sync.RWMutex{}}
	p.SetAPI(mockAPI)
	p.env = Env{httpHandler: handler}

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		p.ServeHTTP(nil, rec, req)
	})

	assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
}

func TestServeHTTPNoPanicSucceeds(t *testing.T) {
	handler := httputils.NewHandler()
	handler.Router.HandleFunc("/ok", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	p := &Plugin{envLock: &sync.RWMutex{}}
	p.env = Env{httpHandler: handler}

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()

	p.ServeHTTP(nil, rec, req)

	assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
}
