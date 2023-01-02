// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"github.com/golang/glog"
	"go.jona.me/crowd"
)

type crowdAuthenticate func(user string, pass string) (crowd.User, error)

type crowdAuth struct {
	crowdAuthenticate crowdAuthenticate
}

func NewCrowdAuth(crowdAuthenticate crowdAuthenticate) Verifier {
	return &crowdAuth{
		crowdAuthenticate: crowdAuthenticate,
	}
}

func (a *crowdAuth) Verify(username UserName, password Password) (bool, error) {
	glog.V(2).Infof("verify user %s with password-length %d", username, len(password))
	_, err := a.crowdAuthenticate(username.String(), password.String())
	if err != nil {
		return false, err
	}
	return true, nil
}
