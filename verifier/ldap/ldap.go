package ldap

import (
	"github.com/bborbe/auth_http_proxy/model"
	"github.com/golang/glog"
	ldap "github.com/jtblin/go-ldap-client"
)

type auth struct {
	client *ldap.LDAPClient
}

func New(
	ldapBase model.LdapBase,
	ldapHost model.LdapHost,
	ldapPort model.LdapPort,
	ldapUseSSL model.LdapUseSSL,
	ldapBindDN model.LdapBindDN,
	ldapBindPassword model.LdapBindPassword,
	ldapUserFilter model.LdapUserFilter,
	ldapGroupFilter model.LdapGroupFilter,
) *auth {
	a := new(auth)
	a.client = &ldap.LDAPClient{
		Base:         ldapBase.String(),
		Host:         ldapHost.String(),
		Port:         ldapPort.Int(),
		UseSSL:       ldapUseSSL.Bool(),
		BindDN:       ldapBindDN.String(),
		BindPassword: ldapBindPassword.String(),
		UserFilter:   ldapUserFilter.String(),
		GroupFilter:  ldapGroupFilter.String(),
	}
	return a
}

func (a *auth) Verify(username model.UserName, password model.Password) (bool, error) {
	glog.V(2).Infof("verify user %s")
	ok, _, err := a.client.Authenticate(username.String(), password.String())
	if err != nil {
		glog.Warningf("authenticate user %v on ldap failed: %v", username, err)
		return false, err
	}
	glog.V(2).Infof("authenticate user %v completed. result: %v", username, ok)
	return ok, nil
}
