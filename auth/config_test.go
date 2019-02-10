package auth

import (
	"testing"

	. "github.com/bborbe/assert"
)

func TestCreateGroupTwo(t *testing.T) {
	groups := CreateGroupsFromString("groupA,groupB")
	if err := AssertThat(len(groups), Is(2)); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(string(groups[0]), Is("groupA")); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(string(groups[1]), Is("groupB")); err != nil {
		t.Fatal(err)
	}
}

func TestCreateGroupNone(t *testing.T) {
	groups := CreateGroupsFromString("")
	if err := AssertThat(len(groups), Is(0)); err != nil {
		t.Fatal(err)
	}
}

func TestParseConfigUserDn(t *testing.T) {
	config, err := ParseConfig([]byte(`{"ldap-user-dn":"foo"}`))
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(config.LdapUserDn.String(), Is("foo")); err != nil {
		t.Fatal(err)
	}
}

func TestParseConfigGroupDn(t *testing.T) {
	config, err := ParseConfig([]byte(`{"ldap-group-dn":"foo"}`))
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(config.LdapGroupDn.String(), Is("foo")); err != nil {
		t.Fatal(err)
	}
}
