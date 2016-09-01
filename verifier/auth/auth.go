package auth

import (
	auth_model "github.com/bborbe/auth/model"
	"github.com/bborbe/auth_http_proxy/model"
	"github.com/bborbe/http/header"
	"github.com/golang/glog"
	"strings"
)

type GroupName string

func (g GroupName) String() string {
	return string(g)
}

func CreateGroupsFromString(groupNames string) []GroupName {
	parts := strings.Split(groupNames, ",")
	groups := make([]GroupName, 0)
	for _, groupName := range parts {
		if len(groupName) > 0 {
			groups = append(groups, GroupName(groupName))
		}
	}
	glog.V(1).Infof("required groups: %v", groups)
	return groups
}

type check func(authToken auth_model.AuthToken, requiredGroups []auth_model.GroupName) (*auth_model.UserName, error)

type auth struct {
	check          check
	requiredGroups []auth_model.GroupName
}

func New(check check, requiredGroups ...GroupName) *auth {
	a := new(auth)
	a.check = check
	a.requiredGroups = []auth_model.GroupName{}
	for _, requiredGroup := range requiredGroups {
		a.requiredGroups = append(a.requiredGroups, auth_model.GroupName(requiredGroup))
	}
	return a
}

func (a *auth) Verify(username model.UserName, password model.Password) (bool, error) {
	glog.V(2).Infof("verify user %s has groups %v", username, a.requiredGroups)
	token := header.CreateAuthorizationToken(username.String(), password.String())
	result, err := a.verify(auth_model.AuthToken(token))
	if err != nil {
		glog.V(1).Infof("verify failed: %v", err)
		return false, err
	}
	glog.V(2).Infof("verify user %s => %v", username, result)
	return result, nil
}

func (a *auth) verify(token auth_model.AuthToken) (bool, error) {
	user, err := a.check(token, a.requiredGroups)
	if err != nil {
		return false, err
	}
	result := len(*user) > 0
	return result, nil
}
