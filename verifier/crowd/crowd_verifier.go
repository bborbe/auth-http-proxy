package crowd

import (
	"github.com/bborbe/auth_http_proxy/model"
	"github.com/golang/glog"
	"go.jona.me/crowd"
)

type crowdAuthenticate func(user string, pass string) (crowd.User, error)

type auth struct {
	crowdAuthenticate crowdAuthenticate
}

func New(crowdAuthenticate crowdAuthenticate) *auth {
	a := new(auth)
	a.crowdAuthenticate = crowdAuthenticate
	return a
}

func (a *auth) Verify(username model.UserName, password model.Password) (bool, error) {
	glog.V(2).Infof("verify user %s with password-length %d", username, len(password))
	_, err := a.crowdAuthenticate(username.String(), password.String())
	if err != nil {
		return false, err
	}
	return true, nil
}
