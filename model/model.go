package model

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/golang/glog"
)

type Config struct {
	Port                    Port                    `json:"port"`
	AuthUrl                 AuthUrl                 `json:"auth-url"`
	AuthApplicationName     AuthApplicationName     `json:"auth-application-name"`
	AuthApplicationPassword AuthApplicationPassword `json:"auth-application-password"`
	TargetAddress           TargetAddress           `json:"target-address"`
	BasicAuthRealm          BasicAuthRealm          `json:"basic-auth-realm"`
	AuthGroups              []AuthGroup             `json:"auth-groups"`
	VerifierType            VerifierType            `json:"verifier"`
	UserFile                UserFile                `json:"file-users"`
	Kind                    Kind                    `json:"kind"`
	LdapBase                LdapBase                `json:"ldap-base"`
	LdapHost                LdapHost                `json:"ldap-host"`
	LdapPort                LdapPort                `json:"ldap-port"`
	LdapUseSSL              LdapUseSSL              `json:"ldap-use-ssl"`
	LdapBindDN              LdapBindDN              `json:"ldap-bind-dn"`
	LdapBindPassword        LdapBindPassword        `json:"ldap-bind-password"`
	LdapUserFilter          LdapUserFilter          `json:"ldap-user-filter"`
	LdapGroupFilter         LdapGroupFilter         `json:"ldap-group-filter"`
}
type ConfigPath string

func (c ConfigPath) IsValue() bool {
	return len(c) > 0
}

func (c ConfigPath) Parse() (*Config, error) {
	content, err := ioutil.ReadFile(string(c))
	if err != nil {
		glog.Warningf("read config from file %v failed: %v", c, err)
		return nil, err
	}
	return ParseConfig(content)
}

func ParseConfig(content []byte) (*Config, error) {
	config := &Config{}
	if err := json.Unmarshal(content, config); err != nil {
		glog.Warningf("parse config failed: %v", err)
		return nil, err
	}
	return config, nil
}

type Port int

type AuthUrl string

func (a AuthUrl) String() string {
	return string(a)
}

type AuthApplicationName string

func (a AuthApplicationName) String() string {
	return string(a)
}

type AuthApplicationPassword string

func (a AuthApplicationPassword) String() string {
	return string(a)
}

type TargetAddress string

func (t TargetAddress) String() string {
	return string(t)
}

type BasicAuthRealm string

func (b BasicAuthRealm) String() string {
	return string(b)
}

type AuthGroup string

func (a AuthGroup) String() string {
	return string(a)
}

type Kind string

func (k Kind) String() string {
	return string(k)
}

type VerifierType string

func (v VerifierType) String() string {
	return string(v)
}

type UserName string

func (u UserName) String() string {
	return string(u)
}

type Password string

func (p Password) String() string {
	return string(p)
}

type UserFile string

func (u UserFile) String() string {
	return string(u)
}

func CreateGroupsFromString(groupNames string) []AuthGroup {
	parts := strings.Split(groupNames, ",")
	groups := make([]AuthGroup, 0)
	for _, groupName := range parts {
		if len(groupName) > 0 {
			groups = append(groups, AuthGroup(groupName))
		}
	}
	glog.V(1).Infof("required groups: %v", groups)
	return groups
}

type LdapBase string

func (l LdapBase) String() string {
	return string(l)
}

type LdapHost string

func (l LdapHost) String() string {
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
