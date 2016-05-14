package auth_verifier

import (
	"github.com/bborbe/auth/api"
	"github.com/bborbe/http/header"
)

type Check func(authToken api.AuthToken, requiredGroups []api.GroupName) (*api.UserName, error)

type auth struct {
	check          Check
	requiredGroups []api.GroupName
}

func New(check Check, requiredGroups []api.GroupName) *auth {
	a := new(auth)
	a.check = check
	a.requiredGroups = requiredGroups
	return a
}

func (a *auth) Verify(username string, password string) (bool, error) {
	token := header.CreateAuthorizationToken(username, password)
	user, err := a.check(api.AuthToken(token), a.requiredGroups)
	if err != nil {
		return false, err
	}
	return len(*user) > 0, nil
}
