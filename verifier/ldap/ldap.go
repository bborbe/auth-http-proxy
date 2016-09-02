package ldap

import (
	"github.com/bborbe/auth_http_proxy/model"
	"github.com/golang/glog"
	ldap "github.com/jtblin/go-ldap-client"
)

type auth struct {
	client         *ldap.LDAPClient
	requiredGroups []model.GroupName
}

func New(
	ldapBase model.LdapBase,
	ldapHost model.LdapHost,
	ldapServerName model.LdapServerName,
	ldapPort model.LdapPort,
	ldapUseSSL model.LdapUseSSL,
	ldapBindDN model.LdapBindDN,
	ldapBindPassword model.LdapBindPassword,
	ldapUserFilter model.LdapUserFilter,
	ldapGroupFilter model.LdapGroupFilter,
	ldapUserDn model.LdapUserDn,
	ldapGroupDn model.LdapGroupDn,
	requiredGroups ...model.GroupName,
) *auth {
	a := new(auth)
	a.client = &ldap.LDAPClient{
		Base:         ldapBase.String(),
		Host:         ldapHost.String(),
		ServerName:   ldapServerName.String(),
		Port:         ldapPort.Int(),
		UseSSL:       ldapUseSSL.Bool(),
		BindDN:       ldapBindDN.String(),
		BindPassword: ldapBindPassword.String(),
		UserFilter:   ldapUserFilter.String(),
		GroupFilter:  ldapGroupFilter.String(),
		UserDN:       ldapUserDn.String(),
		GroupDN:      ldapGroupDn.String(),
	}
	a.requiredGroups = requiredGroups
	return a
}

func (a *auth) Verify(username model.UserName, password model.Password) (bool, error) {
	glog.V(2).Infof("verify user %v is valid and has groups %v", username, a.requiredGroups)
	glog.V(2).Infof("verify username and password of user %v", username)
	ok, _, err := a.client.Authenticate(username.String(), password.String())
	if err != nil {
		glog.Warningf("verify username and password of user %v on ldap failed: %v", username, err)
		return false, err
	}
	if !ok {
		glog.V(1).Infof("authenticate user %v invalid", username)
		return false, nil
	}
	glog.V(2).Infof("username and password of user %v is valid", username)
	glog.V(2).Infof("get groups of user %v", username)
	groupNames, err := a.client.GetGroupsOfUser(username.String())
	if err != nil {
		glog.Warningf("get groups for user %v failed: %v", username, err)
		return false, err
	}
	glog.V(2).Infof("user %v has groups: %v", username, groupNames)
	for _, requiredGroup := range a.requiredGroups {
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
