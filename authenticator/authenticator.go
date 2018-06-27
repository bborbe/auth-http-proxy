package authenticator

import "github.com/bborbe/auth-http-proxy/model"

type Authenticator interface {
	Authenticate(model.UserName, model.Password) (bool, map[string]string, error)
}
