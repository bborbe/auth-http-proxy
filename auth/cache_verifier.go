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

func (a *cacheAuth) Verify(username UserName, password Password) (bool, error) {
	glog.V(2).Infof("verify user %s with password-length %d", username, len(password))
	value, found := a.cache.Get(username.String())
	if found && value == password.String() {
		glog.V(2).Infof("cache hit for user %v", username)
		return true, nil
	}
	result, err := a.Verify(username, password)
	if err != nil {
		glog.Warningf("verify user %v failed: %v", username, err)
		return false, err
	}
	if result {
		glog.V(2).Infof("add user %v to cache", username)
		a.cache.Set(username.String(), password.String())
	}
	return result, nil
}
