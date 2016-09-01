package verifier

import "github.com/bborbe/auth_http_proxy/model"

type Verifier interface {
	Verify(model.UserName, model.Password) (bool, error)
}
