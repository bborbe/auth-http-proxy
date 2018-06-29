package ldap

import (
	"github.com/jtblin/go-ldap-client"
	"github.com/bborbe/auth-http-proxy/model"
)

type Authenticator struct {
	Client *ldap.LDAPClient
}

func (a *Authenticator) Authenticate(username model.UserName, password model.Password) (bool, map[string]string, error) {
	return a.Client.Authenticate(username.String(), password.String())
}
