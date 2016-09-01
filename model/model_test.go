package model

import (
	"testing"

	"os"

	. "github.com/bborbe/assert"
	"github.com/golang/glog"
)

func TestMain(m *testing.M) {
	exit := m.Run()
	glog.Flush()
	os.Exit(exit)
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
