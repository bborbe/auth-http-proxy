// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"github.com/golang/glog"
	"github.com/jtblin/go-ldap-client"
)

const ldapConnectionSize = 5

type LdapBaseDn string

func (l LdapBaseDn) String() string {
	return string(l)
}

type LdapHost string

func (l LdapHost) String() string {
	return string(l)
}

type LdapServerName string

func (l LdapServerName) String() string {
	return string(l)
}

type LdapPort int

func (l LdapPort) Int() int {
	return int(l)
}

type LdapUseSSL bool

func (l LdapUseSSL) Bool() bool {
	return bool(l)
}

type LdapSkipTls bool

func (l LdapSkipTls) Bool() bool {
	return bool(l)
}

type LdapBindDN string

func (l LdapBindDN) String() string {
	return string(l)
}

type LdapBindPassword string

func (l LdapBindPassword) String() string {
	return string(l)
}

type LdapUserFilter string

func (l LdapUserFilter) String() string {
	return string(l)
}

type LdapGroupFilter string

func (l LdapGroupFilter) String() string {
	return string(l)
}

type LdapUserDn string

func (l LdapUserDn) String() string {
	return string(l)
}

type LdapGroupDn string

func (l LdapGroupDn) String() string {
	return string(l)
}

type LdapUserField string

func (l LdapUserField) String() string {
	return string(l)
}

type LdapGroupField string

func (l LdapGroupField) String() string {
	return string(l)
}

type LdapAuthenticator interface {
	Authenticate(UserName, Password) (bool, map[string]string, error)
	GetGroupsOfUser(UserName) ([]string, error)
}

type ldapAuth struct {
	ldapBaseDn       LdapBaseDn
	ldapHost         LdapHost
	ldapServerName   LdapServerName
	ldapPort         LdapPort
	ldapUseSSL       LdapUseSSL
	ldapSkipTls      LdapSkipTls
	ldapBindDN       LdapBindDN
	ldapBindPassword LdapBindPassword
	ldapUserFilter   LdapUserFilter
	ldapGroupFilter  LdapGroupFilter
	ldapUserDn       LdapUserDn
	ldapGroupDn      LdapGroupDn
	ldapUserField    LdapUserField
	ldapGroupField   LdapGroupField
	ldapClients      chan *ldap.LDAPClient
}

func NewLdapAuthenticator(
	ldapBaseDn LdapBaseDn,
	ldapHost LdapHost,
	ldapServerName LdapServerName,
	ldapPort LdapPort,
	ldapUseSSL LdapUseSSL,
	ldapSkipTls LdapSkipTls,
	ldapBindDN LdapBindDN,
	ldapBindPassword LdapBindPassword,
	ldapUserDn LdapUserDn,
	ldapUserFilter LdapUserFilter,
	ldapUserField LdapUserField,
	ldapGroupDn LdapGroupDn,
	ldapGroupFilter LdapGroupFilter,
	ldapGroupField LdapGroupField,
) LdapAuthenticator {
	a := new(ldapAuth)
	a.ldapBaseDn = ldapBaseDn
	a.ldapHost = ldapHost
	a.ldapServerName = ldapServerName
	a.ldapPort = ldapPort
	a.ldapUseSSL = ldapUseSSL
	a.ldapSkipTls = ldapSkipTls
	a.ldapBindDN = ldapBindDN
	a.ldapBindPassword = ldapBindPassword
	a.ldapUserFilter = ldapUserFilter
	a.ldapGroupFilter = ldapGroupFilter
	a.ldapUserField = ldapUserField
	a.ldapGroupField = ldapGroupField
	a.ldapUserDn = ldapUserDn
	a.ldapGroupDn = ldapGroupDn
	a.ldapClients = make(chan *ldap.LDAPClient, ldapConnectionSize)
	return a
}

func (a *ldapAuth) getClient() *ldap.LDAPClient {
	select {
	case client := <-a.ldapClients:
		glog.V(2).Infof("got client from pool")
		return client
	default:
		glog.V(2).Infof("created new client")
		return a.createClient()
	}
}

func (a *ldapAuth) releaseClient(client *ldap.LDAPClient) {
	glog.V(2).Infof("release client")
	select {
	case a.ldapClients <- client:
		glog.V(2).Infof("returned client to pool")
	default:
		a.closeClient(client)
	}
}

func (a *ldapAuth) closeClient(client *ldap.LDAPClient) {
	glog.V(2).Infof("closed client")
	client.Close()
}

func (a *ldapAuth) Close() {
	glog.V(2).Infof("close all ldap connections")
	for client := range a.ldapClients {
		client.Close()
	}
}

func (a *ldapAuth) Authenticate(username UserName, password Password) (ok bool, data map[string]string, err error) {
	glog.V(2).Infof("Authenticate user %s", username)
	ldapClient := a.getClient()
	ok, data, err = ldapClient.Authenticate(username.String(), password.String())
	if err != nil {
		glog.V(1).Infof("Authenticate failed, retry with new connection: %v", err)
		a.closeClient(ldapClient)
		ldapClient = a.createClient()
		ok, data, err = ldapClient.Authenticate(username.String(), password.String())
	}
	a.releaseClient(ldapClient)
	return
}

func (a *ldapAuth) GetGroupsOfUser(username UserName) (groups []string, err error) {
	glog.V(2).Infof("GetGroupsOfUser for user %s", username)
	ldapClient := a.getClient()
	groups, err = ldapClient.GetGroupsOfUser(username.String())
	if err != nil {
		glog.V(1).Infof("GetGroupsOfUser failed, retry with new connection: %v", err)
		a.closeClient(ldapClient)
		ldapClient = a.createClient()
		groups, err = ldapClient.GetGroupsOfUser(username.String())
	}
	a.releaseClient(ldapClient)
	return
}

func (a *ldapAuth) createClient() *ldap.LDAPClient {
	serverName := a.ldapServerName.String()
	if len(serverName) == 0 {
		serverName = a.ldapHost.String()
	}
	glog.V(2).Infof("create new ldap client for %s:%d with servername %s", a.ldapHost, a.ldapPort, serverName)
	client := &ldap.LDAPClient{
		Base:         a.ldapBaseDn.String(),
		BindDN:       a.ldapBindDN.String(),
		BindPassword: a.ldapBindPassword.String(),
		GroupDN:      a.ldapGroupDn.String(),
		GroupField:   a.ldapGroupField.String(),
		GroupFilter:  a.ldapGroupFilter.String(),
		Host:         a.ldapHost.String(),
		Port:         a.ldapPort.Int(),
		ServerName:   serverName,
		SkipTLS:      a.ldapSkipTls.Bool(),
		UseSSL:       a.ldapUseSSL.Bool(),
		UserDN:       a.ldapUserDn.String(),
		UserField:    a.ldapUserField.String(),
		UserFilter:   a.ldapUserFilter.String(),
	}
	if glog.V(4) {
		glog.Infof("client %+v", client)
	}
	return client
}
