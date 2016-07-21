package auth_verifier

import (
	"testing"

	"fmt"

	. "github.com/bborbe/assert"
	"github.com/bborbe/auth/api"
)

func TestVerifyFailed(t *testing.T) {
	username := "user123"
	password := "pass123"

	authVerifier := New(func(authToken api.AuthToken, requiredGroups []api.GroupName) (*api.UserName, error) {
		return nil, fmt.Errorf("not found")
	}, api.GroupName("test"))

	result, err := authVerifier.Verify(username, password)
	if err := AssertThat(err, NotNilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(result, Is(false)); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyNotFound(t *testing.T) {
	username := "user123"
	password := "pass123"

	authVerifier := New(func(authToken api.AuthToken, requiredGroups []api.GroupName) (*api.UserName, error) {
		u := api.UserName("")
		return &u, nil
	}, api.GroupName("test"))

	result, err := authVerifier.Verify(username, password)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(result, Is(false)); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyFound(t *testing.T) {
	username := "user123"
	password := "pass123"

	authVerifier := New(func(authToken api.AuthToken, requiredGroups []api.GroupName) (*api.UserName, error) {
		u := api.UserName(username)
		return &u, nil
	}, api.GroupName("test"))

	result, err := authVerifier.Verify(username, password)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(result, Is(true)); err != nil {
		t.Fatal(err)
	}
}
