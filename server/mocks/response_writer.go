package mocks

import "net/http"

var _ http.ResponseWriter = &MockResponseWriter{}

func DefaultMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		HeaderMap:  make(http.Header),
		Bytes:      make([]byte, 0),
		Err:        nil,
		StatusCode: http.StatusOK,
	}
}

type MockResponseWriter struct {
	HeaderMap  http.Header
	Bytes      []byte
	Err        error
	StatusCode int
}

func (rw *MockResponseWriter) Header() http.Header {
	return rw.HeaderMap
}

func (rw *MockResponseWriter) Write(bytes []byte) (int, error) {
	if rw.Err != nil {
		return 0, rw.Err
	}

	rw.Bytes = append(rw.Bytes, bytes...)

	return len(rw.Bytes), nil
}

func (rw *MockResponseWriter) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
}
