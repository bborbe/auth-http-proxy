package auth

import (
	"testing"

	"fmt"

	"os"

	. "github.com/bborbe/assert"
	auth_model "github.com/bborbe/auth/model"
	"github.com/bborbe/auth_http_proxy/model"
	"github.com/bborbe/auth_http_proxy/verifier"
	"github.com/golang/glog"
)

func TestMain(m *testing.M) {
	exit := m.Run()
	glog.Flush()
	os.Exit(exit)
}

func TestImplementsVerifier(t *testing.T) {
	object := New(nil)
	var expected *verifier.Verifier
	err := AssertThat(object, Implements(expected))
	if err != nil {
		t.Fatal(err)
	}
}

func TestVerifyFailed(t *testing.T) {
	username := model.UserName("user123")
	password := model.Password("pass123")

	authVerifier := New(func(authToken auth_model.AuthToken, requiredGroups []auth_model.GroupName) (*auth_model.UserName, error) {
		return nil, fmt.Errorf("not found")
	}, GroupName("test"))

	result, err := authVerifier.Verify(username, password)
	if err := AssertThat(err, NotNilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(result, Is(false)); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyNotFound(t *testing.T) {
	username := model.UserName("user123")
	password := model.Password("pass123")

	authVerifier := New(func(authToken auth_model.AuthToken, requiredGroups []auth_model.GroupName) (*auth_model.UserName, error) {
		u := auth_model.UserName("")
		return &u, nil
	}, GroupName("test"))

	result, err := authVerifier.Verify(username, password)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(result, Is(false)); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyFound(t *testing.T) {
	username := model.UserName("user123")
	password := model.Password("pass123")

	authVerifier := New(func(authToken auth_model.AuthToken, requiredGroups []auth_model.GroupName) (*auth_model.UserName, error) {
		u := auth_model.UserName(username)
		return &u, nil
	}, GroupName("test"))

	result, err := authVerifier.Verify(username, password)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(result, Is(true)); err != nil {
		t.Fatal(err)
	}
}

func TestCreateGroupOne(t *testing.T) {
	groups := CreateGroupsFromString("test")
	if err := AssertThat(len(groups), Is(1)); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(string(groups[0]), Is("test")); err != nil {
		t.Fatal(err)
	}
}

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
