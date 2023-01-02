// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package auth

import (
	"time"

	"github.com/golang/glog"
	"github.com/wunderlist/ttlcache"
)

type CacheTTL time.Duration

func (c CacheTTL) IsEmpty() bool {
	return int64(c) == 0
}

func (c CacheTTL) Duration() time.Duration {
	return time.Duration(c)
}

type cacheAuth struct {
	verifier Verifier
	cache    *ttlcache.Cache
}

func NewCacheAuth(
	verifier Verifier,
	ttl CacheTTL,
) Verifier {
	return &cacheAuth{
		verifier: verifier,
		cache:    ttlcache.NewCache(ttl.Duration()),
	}
}

func (c *cacheAuth) Verify(username UserName, password Password) (bool, error) {
	glog.V(2).Infof("verify user %s with password-length %d", username, len(password))
	value, found := c.cache.Get(username.String())
	if found && value == password.String() {
		glog.V(2).Infof("cache hit for user %v", username)
		return true, nil
	}
	result, err := c.verifier.Verify(username, password)
	if err != nil {
		glog.Warningf("verify user %v failed: %v", username, err)
		return false, err
	}
	if result {
		glog.V(2).Infof("add user %v to cache", username)
		c.cache.Set(username.String(), password.String())
	}
	return result, nil
}
