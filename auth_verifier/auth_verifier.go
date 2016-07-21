package auth_verifier

import (
	"github.com/bborbe/auth/api"
	"github.com/bborbe/http/header"
	"github.com/bborbe/log"
)

var logger = log.DefaultLogger

type Check func(authToken api.AuthToken, requiredGroups []api.GroupName) (*api.UserName, error)

type auth struct {
	check          Check
	requiredGroups []api.GroupName
}

func New(check Check, requiredGroups ...api.GroupName) *auth {
	a := new(auth)
	a.check = check
	a.requiredGroups = requiredGroups
	return a
}

func (a *auth) Verify(username string, password string) (bool, error) {
	logger.Debugf("verify user %s has groups %v", username, a.requiredGroups)
	token := header.CreateAuthorizationToken(username, password)
	user, err := a.check(api.AuthToken(token), a.requiredGroups)
	if err != nil {
		logger.Debugf("verify failed: %v", err)
		return false, err
	}
	result := len(*user) > 0
	logger.Debugf("verify user %s => %v", username, result)
	return result, nil
}
