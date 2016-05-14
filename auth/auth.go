package auth

import (
	"net/http"

	"github.com/bborbe/auth/api"
	"github.com/bborbe/http/header"
	"github.com/bborbe/log"
)

var logger = log.DefaultLogger

const REQUIRED_GROUP = api.GroupName("aptly-api")

type Auth func(authToken api.AuthToken, requiredGroups []api.GroupName) (*api.UserName, error)

type handler struct {
	auth    Auth
	handler http.Handler
}

func New(subhandler http.Handler, auth Auth) *handler {
	h := new(handler)
	h.handler = subhandler
	h.auth = auth
	return h
}

func (h *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	logger.Debugf("auth handler")
	if err := h.serveHTTP(resp, req); err != nil {
		status := http.StatusUnauthorized
		logger.Debugf("auth failed => send %s", http.StatusText(status))
		resp.WriteHeader(status)
	} else {
		h.handler.ServeHTTP(resp, req)
	}
}

func (h *handler) serveHTTP(resp http.ResponseWriter, req *http.Request) error {
	name, value, err := header.ParseAuthorizationBasisHttpRequest(req)
	if err != nil {
		return err
	}
	token := header.CreateAuthorizationToken(name, value)
	logger.Debugf("token: %s", token)
	user, err := h.auth(api.AuthToken(token), []api.GroupName{REQUIRED_GROUP})
	if err != nil {
		logger.Debugf("get user with token %s and group %v faild", token, REQUIRED_GROUP)
		return err
	}
	logger.Debugf("user %v is in group %v", *user, REQUIRED_GROUP)
	return nil
}
