package ldap

import (
	"github.com/bborbe/auth_http_proxy/model"
	"github.com/golang/glog"
	ldap "github.com/jtblin/go-ldap-client"
)

const ldapConnectionSize = 5

type auth struct {
	ldapBase         model.LdapBase
	ldapHost         model.LdapHost
	ldapServerName   model.LdapServerName
	ldapPort         model.LdapPort
	ldapUseSSL       model.LdapUseSSL
	ldapBindDN       model.LdapBindDN
	ldapBindPassword model.LdapBindPassword
	ldapUserFilter   model.LdapUserFilter
	ldapGroupFilter  model.LdapGroupFilter
	ldapUserDn       model.LdapUserDn
	ldapGroupDn      model.LdapGroupDn
	requiredGroups   []model.GroupName
	ldapClients      chan *ldap.LDAPClient
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
	a.ldapBase = ldapBase
	a.ldapHost = ldapHost
	a.ldapServerName = ldapServerName
	a.ldapPort = ldapPort
	a.ldapUseSSL = ldapUseSSL
	a.ldapBindDN = ldapBindDN
	a.ldapBindPassword = ldapBindPassword
	a.ldapUserFilter = ldapUserFilter
	a.ldapGroupFilter = ldapGroupFilter
	a.ldapUserDn = ldapUserDn
	a.ldapGroupDn = ldapGroupDn
	a.requiredGroups = requiredGroups
	a.ldapClients = make(chan *ldap.LDAPClient, ldapConnectionSize)
	return a
}

func (a *auth) createClient() *ldap.LDAPClient {
	return &ldap.LDAPClient{
		Base:         a.ldapBase.String(),
		Host:         a.ldapHost.String(),
		ServerName:   a.ldapServerName.String(),
		Port:         a.ldapPort.Int(),
		UseSSL:       a.ldapUseSSL.Bool(),
		BindDN:       a.ldapBindDN.String(),
		BindPassword: a.ldapBindPassword.String(),
		UserFilter:   a.ldapUserFilter.String(),
		GroupFilter:  a.ldapGroupFilter.String(),
		UserDN:       a.ldapUserDn.String(),
		GroupDN:      a.ldapGroupDn.String(),
	}
}

func (a *auth) getClient() *ldap.LDAPClient {
	select {
	case client := <-a.ldapClients:
		glog.V(2).Infof("got client from pool")
		return client
	default:
		glog.V(2).Infof("created new client")
		return a.createClient()
	}
}

func (a *auth) releaseClient(client *ldap.LDAPClient) {
	select {
	case a.ldapClients <- client:
		glog.V(2).Infof("returned client to pool")
	default:
		glog.V(2).Infof("closed client")
		client.Close()
	}
}

func (a *auth) Close() {
	glog.V(2).Infof("close all ldap connections")
	for client := range a.ldapClients {
		client.Close()
	}
}

func (a *auth) Verify(username model.UserName, password model.Password) (bool, error) {
	client := a.getClient()
	defer a.releaseClient(client)

	glog.V(2).Infof("verify user %v is valid and has groups %v", username, a.requiredGroups)
	glog.V(2).Infof("verify username and password of user %v", username)
	ok, _, err := client.Authenticate(username.String(), password.String())
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
	groupNames, err := client.GetGroupsOfUser(username.String())
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
