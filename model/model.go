package model

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"time"

	"fmt"

	"github.com/golang/glog"
)

type Config struct {
	Port                    Port                    `json:"port"`
	CacheTTL                CacheTTL                `json:"cache-ttl"`
	AuthUrl                 AuthUrl                 `json:"auth-url"`
	AuthApplicationName     AuthApplicationName     `json:"auth-application-name"`
	AuthApplicationPassword AuthApplicationPassword `json:"auth-application-password"`
	TargetAddress           TargetAddress           `json:"target-address"`
	TargetHealthzUrl        TargetHealthzUrl        `json:"target-healthz-url"`
	BasicAuthRealm          BasicAuthRealm          `json:"basic-auth-realm"`
	Secret                  Secret                  `json:"secret"`
	RequiredGroups          []GroupName             `json:"required-groups"`
	VerifierType            VerifierType            `json:"verifier"`
	UserFile                UserFile                `json:"file-users"`
	Kind                    Kind                    `json:"kind"`
	LdapBase                LdapBase                `json:"ldap-base"`
	LdapHost                LdapHost                `json:"ldap-host"`
	LdapServerName          LdapServerName          `json:"ldap-servername"`
	LdapPort                LdapPort                `json:"ldap-port"`
	LdapUseSSL              LdapUseSSL              `json:"ldap-use-ssl"`
	LdapBindDN              LdapBindDN              `json:"ldap-bind-dn"`
	LdapBindPassword        LdapBindPassword        `json:"ldap-bind-password"`
	LdapUserFilter          LdapUserFilter          `json:"ldap-user-filter"`
	LdapGroupFilter         LdapGroupFilter         `json:"ldap-group-filter"`
	LdapUserDn              LdapUserDn              `json:"ldap-user-dn"`
	LdapGroupDn             LdapGroupDn             `json:"ldap-group-dn"`
	CrowdURL                CrowdURL                `json:"crowd-url"`
	CrowdAppName            CrowdAppName            `json:"crowd-app-name"`
	CrowdAppPassword        CrowdAppPassword        `json:"crowd-app-password"`
}

type CacheTTL time.Duration

func (c CacheTTL) IsEmpty() bool {
	return int64(c) == 0
}

func (c CacheTTL) Duration() time.Duration {
	return time.Duration(c)
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

func (p Port) Address() string {
	return fmt.Sprintf(":%d", p)
}

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

type TargetHealthzUrl string

func (t TargetHealthzUrl) String() string {
	return string(t)
}

type BasicAuthRealm string

func (b BasicAuthRealm) String() string {
	return string(b)
}

type Secret string

func (s Secret) String() string {
	return string(s)
}

func (s Secret) Bytes() []byte {
	return []byte(s)
}

type GroupName string

func (g GroupName) String() string {
	return string(g)
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

func CreateGroupsFromString(groupNames string) []GroupName {
	parts := strings.Split(groupNames, ",")
	groups := make([]GroupName, 0)
	for _, groupName := range parts {
		if len(groupName) > 0 {
			groups = append(groups, GroupName(groupName))
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

type CrowdURL string

func (c CrowdURL) String() string {
	return string(c)
}

type CrowdAppName string

func (c CrowdAppName) String() string {
	return string(c)
}

type CrowdAppPassword string

func (c CrowdAppPassword) String() string {
	return string(c)
}
