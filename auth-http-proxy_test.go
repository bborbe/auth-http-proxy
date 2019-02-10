package main

import (
	"testing"

	. "github.com/bborbe/assert"
)

func TestCreateConfig(t *testing.T) {
	config, err := createConfig()
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(config, NotNilValue()); err != nil {
		t.Fatal(err)
	}

}
