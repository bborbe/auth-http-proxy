// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"github.com/golang/glog"
)

type GroupName string

func (g GroupName) String() string {
	return string(g)
}

type LdapAuth struct {
	LdapAuthenticator LdapAuthenticator
	RequiredGroups    []GroupName
}

func (l *LdapAuth) Verify(username UserName, password Password) (bool, error) {
	glog.V(2).Infof("verify user %v is valid and has groups %v", username, l.RequiredGroups)

	ok, _, err := l.LdapAuthenticator.Authenticate(username, password)
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
	groupNames, err := l.LdapAuthenticator.GetGroupsOfUser(username)
	if err != nil {
		glog.Warningf("get groups for user %v failed: %v", username, err)
		return false, err
	}
	glog.V(2).Infof("user %v has groups: %v", username, groupNames)
	for _, requiredGroup := range l.RequiredGroups {
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
