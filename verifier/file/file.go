package file

import (
	"github.com/bborbe/auth_http_proxy/model"
	"github.com/golang/glog"
)

type auth struct {
}

func New() *auth {
	a := new(auth)
	return a
}

func (a *auth) Verify(username model.UserName, password model.Password) (bool, error) {
	glog.V(2).Infof("verify user %s")
	return false, nil
}
