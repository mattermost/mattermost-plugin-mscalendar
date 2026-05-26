// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package httputils

import (
	"net/http"

	"github.com/gorilla/mux"
)

const maxRequestBodyBytes int64 = 1 << 20 // 1 MB

type Handler struct {
	root *mux.Router

	// Router is the mux.Router that sub-packages register routes on.
	// It sits behind the global middleware applied in ServeHTTP.
	*mux.Router
}

func NewHandler() *Handler {
	router := mux.NewRouter()
	router.Handle("{anything:.*}", http.NotFoundHandler())
	return &Handler{
		root:   router,
		Router: router,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	h.root.ServeHTTP(w, r)
}
