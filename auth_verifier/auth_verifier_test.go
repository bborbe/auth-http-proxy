package auth_verifier

import (
	"testing"

	"fmt"

	. "github.com/bborbe/assert"
	"github.com/bborbe/auth/model"
	"github.com/golang/glog"
	"os"
)

func TestMain(m *testing.M) {
	exit := m.Run()
	glog.Flush()
	os.Exit(exit)
}

func TestVerifyFailed(t *testing.T) {
	username := "user123"
	password := "pass123"

	authVerifier := New(func(authToken model.AuthToken, requiredGroups []model.GroupName) (*model.UserName, error) {
		return nil, fmt.Errorf("not found")
	}, model.GroupName("test"))

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

	authVerifier := New(func(authToken model.AuthToken, requiredGroups []model.GroupName) (*model.UserName, error) {
		u := model.UserName("")
		return &u, nil
	}, model.GroupName("test"))

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

	authVerifier := New(func(authToken model.AuthToken, requiredGroups []model.GroupName) (*model.UserName, error) {
		u := model.UserName(username)
		return &u, nil
	}, model.GroupName("test"))

	result, err := authVerifier.Verify(username, password)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(result, Is(true)); err != nil {
		t.Fatal(err)
	}
}
