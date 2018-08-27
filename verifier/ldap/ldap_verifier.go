package ldap

import (
	"github.com/bborbe/auth-http-proxy/model"
	"github.com/golang/glog"
	authenticator "github.com/bborbe/auth-http-proxy/ldap"
)

type Auth struct {
	Authenticator  authenticator.Authenticator
	RequiredGroups []model.GroupName
}

func (a *Auth) Verify(username model.UserName, password model.Password) (bool, error) {
	glog.V(2).Infof("verify user %v is valid and has groups %v", username, a.RequiredGroups)

	ok, _, err := a.Authenticator.Authenticate(username, password)
	if err != nil {
		glog.V(0).Infof("authenticate user %v failed %v", username, err)
		return false, nil
	}
	if !ok {
		glog.V(1).Infof("authenticate user %v invalid", username)
		return false, nil
	}

	glog.V(2).Infof("username and password of user %v is valid", username)
	glog.V(2).Infof("get groups of user %v", username)
	groupNames, err := a.Authenticator.GetGroupsOfUser(username)
	if err != nil {
		glog.Warningf("get groups for user %v failed: %v", username, err)
		return false, err
	}
	glog.V(2).Infof("user %v has groups: %v", username, groupNames)
	for _, requiredGroup := range a.RequiredGroups {
		found := false
		for _, groupName := range groupNames {
			if groupName == requiredGroup.String() {
				found = true
			}
		}
		if !found {
			glog.V(1).Infof("user %v has not required group %v", username, requiredGroup)
			return false, nil
		}
	}
	glog.V(2).Infof("user %v is valid and has all required groups", username)
	return true, nil
}
