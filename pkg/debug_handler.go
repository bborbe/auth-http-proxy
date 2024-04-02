// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"net/http"
	"time"

	"github.com/golang/glog"
)

type debugHandler struct {
	subhandler http.Handler
}

func NewDebugHandler(subhandler http.Handler) *debugHandler {
	h := new(debugHandler)
	h.subhandler = subhandler
	return h
}

func (h *debugHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	start := time.Now()
	defer glog.V(4).Infof("%s %s takes %dms", request.Method, request.URL.Path, time.Now().Sub(start)/time.Millisecond)

	glog.V(4).Infof("request %v: ", request)
	h.subhandler.ServeHTTP(responseWriter, request)
	glog.V(4).Infof("response %v: ", responseWriter)
}
