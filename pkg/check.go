// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
)

type checkHandler struct {
	check func() error
}

func NewCheckHandler(c func() error) *checkHandler {
	h := new(checkHandler)
	h.check = c
	return h
}

func (h *checkHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if err := h.check(); err != nil {
		glog.V(2).Infof("check => failed: %v", err)
		http.Error(resp, fmt.Sprintf("check failed"), http.StatusInternalServerError)
		return
	}
	glog.V(4).Infof("check => ok")
	resp.WriteHeader(http.StatusOK)
	fmt.Fprintf(resp, "ok")
}
