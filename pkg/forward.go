// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"fmt"
	"io"
	"net/http"

	"github.com/golang/glog"
)

type executeRequest func(address string, req *http.Request) (resp *http.Response, err error)

type forwardHandler struct {
	target         string
	executeRequest executeRequest
}

func NewForwardHandler(target string, executeRequest executeRequest) *forwardHandler {
	h := new(forwardHandler)
	h.target = target
	h.executeRequest = executeRequest
	return h
}

func (h *forwardHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	glog.V(4).Infof("forward request")
	if err := h.serveHTTP(resp, req); err != nil {
		glog.V(2).Infof("forward request failed: %v", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	glog.V(4).Infof("request forward succesful")
}

func (h *forwardHandler) serveHTTP(resp http.ResponseWriter, req *http.Request) error {
	glog.V(4).Infof("%v", req)
	urlStr := fmt.Sprintf("http://%s%s", req.Host, req.RequestURI)
	glog.V(4).Infof("forward request %s %s", req.Method, urlStr)
	subreq, err := http.NewRequest(req.Method, urlStr, req.Body)
	if err != nil {
		glog.V(2).Infof("create request to %v failed: %v", urlStr, err)
		return err
	}
	subreq.Header = req.Header
	subresp, err := h.executeRequest(h.target, subreq)
	if err != nil {
		glog.V(2).Infof("execute request to %v failed: %v", h.target, err)
		return err
	}
	glog.V(4).Infof("write response")
	copyHeader(resp, &subresp.Header)
	resp.WriteHeader(subresp.StatusCode)
	if _, err := io.Copy(resp, subresp.Body); err != nil {
		glog.V(2).Infof("copy body failed: %v", err)
		return err
	}
	glog.V(4).Infof("forward request done")
	return nil
}

func copyHeader(resp http.ResponseWriter, req *http.Header) {
	for key, values := range *req {
		for _, value := range values {
			resp.Header().Add(key, value)
		}
	}
}
