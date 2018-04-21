package auth_basic

import (
	"fmt"
	"net/http"

	"github.com/bborbe/http/header"
	"github.com/golang/glog"
)

type Check func(username string, password string) (bool, error)

type handler struct {
	handler http.HandlerFunc
	check   Check
	realm   string
}

func New(subhandler http.HandlerFunc, check Check, realm string) *handler {
	h := new(handler)
	h.handler = subhandler
	h.check = check
	h.realm = realm
	return h
}

func (h *handler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	glog.V(4).Infof("check basic auth")
	if err := h.serveHTTP(responseWriter, request); err != nil {
		responseWriter.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", h.realm))
		responseWriter.WriteHeader(http.StatusUnauthorized)
	}
}

func (h *handler) serveHTTP(responseWriter http.ResponseWriter, request *http.Request) error {
	glog.V(4).Infof("check basic auth")
	user, pass, err := header.ParseAuthorizationBasisHttpRequest(request)
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
