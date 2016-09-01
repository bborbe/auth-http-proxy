package main

import (
	"github.com/bborbe/auth_http_proxy/model"
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

func TestGetVerifierByType(t *testing.T) {
	_, err := getVerifierByType(&model.Config{})
	if err := AssertThat(err, NotNilValue()); err != nil {
		t.Fatal(err)
	}
}
