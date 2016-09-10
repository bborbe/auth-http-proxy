package main

import (
	"testing"

	"github.com/bborbe/auth_http_proxy/model"

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
	_, err := createVerifier(&model.Config{})
	if err := AssertThat(err, NotNilValue()); err != nil {
		t.Fatal(err)
	}
}

func TestCreateConfig(t *testing.T) {
	config, err := createConfig()
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(config, NotNilValue()); err != nil {
		t.Fatal(err)
	}

}
