package verifier

import "github.com/bborbe/auth-http-proxy/model"

type Verifier interface {
	Verify(model.UserName, model.Password) (bool, error)
}
