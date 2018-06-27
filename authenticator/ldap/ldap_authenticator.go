package ldap

import (
	"github.com/jtblin/go-ldap-client"
	"github.com/bborbe/auth-http-proxy/model"
)

type Authenticator struct {
	client *ldap.LDAPClient
}

func New(client *ldap.LDAPClient) *Authenticator {
	a := new(Authenticator)
	a.client = client
	return a
}

func (a *Authenticator) Authenticate(username model.UserName, password model.Password) (bool, map[string]string, error) {
	return a.client.Authenticate(username.String(), password.String())
}