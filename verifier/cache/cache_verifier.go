package cache

import (
	"github.com/bborbe/auth-http-proxy/model"
	"github.com/bborbe/auth-http-proxy/verifier"
	"github.com/golang/glog"
	"github.com/wunderlist/ttlcache"
)

type auth struct {
	verifier verifier.Verifier
	cache    *ttlcache.Cache
}

func New(
	verifier verifier.Verifier,
	ttl model.CacheTTL,
) *auth {
	a := new(auth)
	a.verifier = verifier
	a.cache = ttlcache.NewCache(ttl.Duration())
	return a
}

func (a *auth) Verify(username model.UserName, password model.Password) (bool, error) {
	glog.V(2).Infof("verify user %s with password-length %d", username, len(password))
	value, found := a.cache.Get(username.String())
	if found && value == password.String() {
		glog.V(2).Infof("cache hit for user %v", username)
		return true, nil
	}
	result, err := a.verifier.Verify(username, password)
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
