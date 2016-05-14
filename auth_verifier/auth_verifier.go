package auth_verifier

import (
	"github.com/bborbe/auth/api"
	"github.com/bborbe/http/header"
)

type Check func(authToken api.AuthToken, requiredGroups []api.GroupName) (*api.UserName, error)

type auth struct {
	check Check
}

func New(check Check) *auth {
	a := new(auth)
	a.check = check
	return a
}

func (a *auth) Verify(username string, password string) (bool, error) {
	token := header.CreateAuthorizationToken(username, password)
	user, err := a.check(api.AuthToken(token), make([]api.GroupName, 0))
	if err != nil {
		return false, err
	}
	return len(*user) > 0, nil
}
