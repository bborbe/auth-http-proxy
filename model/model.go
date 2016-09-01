package model

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/golang/glog"
)

type Port int
type Debug bool
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

type Config struct {
	Port                    Port                    `json:"port"`
	AuthUrl                 AuthUrl                 `json:"auth_url"`
	AuthApplicationName     AuthApplicationName     `json:"auth_application_name"`
	AuthApplicationPassword AuthApplicationPassword `json:"auth_application_password"`
	TargetAddress           TargetAddress           `json:"target_address"`
	BasicAuthRealm          BasicAuthRealm          `json:"basic_auth_realm"`
	AuthGroups              []AuthGroup             `json:"auth_groups"`
	Debug                   Debug                   `json:"debug"`
	VerifierType            VerifierType            `json:"verifier"`
	UserFile                UserFile                `json:"file_users"`
	Kind                    Kind                    `json:"kind"`
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
