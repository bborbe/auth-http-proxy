package cache

import (
	"github.com/bborbe/auth_http_proxy/model"
	"github.com/bborbe/auth_http_proxy/verifier"
	"github.com/golang/glog"
)

type auth struct {
	verifier verifier.Verifier
}

func New(verifier verifier.Verifier) *auth {
	a := new(auth)
	a.verifier = verifier
	return a
}

func (a *auth) Verify(username model.UserName, password model.Password) (bool, error) {
	glog.V(2).Infof("verify user %s with password-length %d", username, len(password))
	return a.verifier.Verify(username, password)
}
