package main

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

func TestCreateGroupOne(t *testing.T) {
	groups := createGroups("test")
	if err := AssertThat(len(groups), Is(1)); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(string(groups[0]), Is("test")); err != nil {
		t.Fatal(err)
	}
}

func TestCreateGroupTwo(t *testing.T) {
	groups := createGroups("groupA,groupB")
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
	groups := createGroups("")
	if err := AssertThat(len(groups), Is(0)); err != nil {
		t.Fatal(err)
	}
}
