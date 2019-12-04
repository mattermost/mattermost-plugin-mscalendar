package http

import "net/http"

var (
	_ http.ResponseWriter = &mockResponseWriter{}
)

func defaultMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{
		HeaderMap:  make(http.Header),
		Bytes:      make([]byte, 0),
		Err:        nil,
		StatusCode: http.StatusOK,
	}
}

type mockResponseWriter struct {
	HeaderMap  http.Header
	Bytes      []byte
	Err        error
	StatusCode int
}

func (rw *mockResponseWriter) Header() http.Header {
	return rw.HeaderMap
}

func (rw *mockResponseWriter) Write(bytes []byte) (int, error) {
	if rw.Err != nil {
		return 0, rw.Err
	}

	rw.Bytes = append(rw.Bytes, bytes...)

	return len(rw.Bytes), nil
}

func (rw *mockResponseWriter) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
}
