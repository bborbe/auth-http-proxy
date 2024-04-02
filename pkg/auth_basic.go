// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
)

func NewAuthBasicHandler(subhandler http.HandlerFunc, check Check, realm string) *authHandler {
	h := new(authHandler)
	h.handler = subhandler
	h.check = check
	h.realm = realm
	return h
}

type authHandler struct {
	handler http.HandlerFunc
	check   Check
	realm   string
}

func (h *authHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	glog.V(4).Infof("check basic auth")
	if err := h.serveHTTP(responseWriter, request); err != nil {
		responseWriter.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", h.realm))
		responseWriter.WriteHeader(http.StatusUnauthorized)
	}
}

func (h *authHandler) serveHTTP(responseWriter http.ResponseWriter, request *http.Request) error {
	glog.V(4).Infof("check basic auth")
	user, pass, err := ParseAuthorizationBasisHttpRequest(request)
	if err != nil {
		glog.Warningf("parse header failed: %v", err)
		return err
	}
	valid, err := h.check(user, pass)
	if err != nil {
		glog.Warningf("check auth for user %v failed: %v", user, err)
		return err
	}
	if !valid {
		glog.V(2).Infof("auth invalid for user %v", user)
		return fmt.Errorf("auth invalid for user %v", user)
	}
	h.handler(responseWriter, request)
	return nil
}
