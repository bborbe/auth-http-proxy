// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
)

func NewAuthBasicHandler(
	subhandler http.Handler,
	check Check,
	realm string,
) http.Handler {
	h := new(authBasicHandler)
	h.handler = subhandler
	h.check = check
	h.realm = realm
	return h
}

type authBasicHandler struct {
	handler http.Handler
	check   Check
	realm   string
}

func (a *authBasicHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	glog.V(4).Infof("check basic auth")
	if err := a.serveHTTP(responseWriter, request); err != nil {
		responseWriter.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", a.realm))
		responseWriter.WriteHeader(http.StatusUnauthorized)
	}
}

func (a *authBasicHandler) serveHTTP(responseWriter http.ResponseWriter, request *http.Request) error {
	glog.V(4).Infof("check basic auth")
	user, pass, err := ParseAuthorizationBasisHttpRequest(request)
	if err != nil {
		glog.Warningf("parse header failed: %v", err)
		return err
	}
	valid, err := a.check.Check(user, pass)
	if err != nil {
		glog.Warningf("check auth for user %v failed: %v", user, err)
		return err
	}
	if !valid {
		glog.V(2).Infof("auth invalid for user %v", user)
		return fmt.Errorf("auth invalid for user %v", user)
	}
	request.Header.Add(ForwardForUserHeader, user)
	a.handler.ServeHTTP(responseWriter, request)
	return nil
}
